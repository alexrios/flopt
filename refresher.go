package flopt

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)
ffc

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Refresher is a cache refresher.
// log is the logger.
// flags is a map of flag keys to their values.
// fetchFn is a function that takes a flag key and returns the new value for that flag.
type Refresher struct {
	log     *zap.Logger
	fetchFn func(context.Context, string) (bool, error)
	// A negative value indicates no limit.
	fetchMaxConcurrency       int
	flags                     *Flags
	failedRefreshes           *prometheus.CounterVec
	failedAttemptsFn          func()
	failedAttemptsMaxDuration time.Duration
	failedAttemptsMaxCount    int32
}

// NewRefresher returns a new Refresher.
func NewRefresher(log *zap.Logger, flags *Flags, fetchFn func(context.Context, string) (bool, error), options ...RefresherOption) (*Refresher, error) {
	if flags == nil {
		return nil, errors.New("flags cannot be nil")
	}
	refresher := Refresher{
		log:                 log,
		fetchFn:             fetchFn,
		fetchMaxConcurrency: -1,
		flags:               flags,
		failedRefreshes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "ff_client_failed_refreshes_total",
			Help: "The total number of failed refreshes",
		},
			[]string{"key"},
		),
	}

	for _, o := range options {
		o(&refresher)
	}

	return &refresher, nil
}

type RefresherOption func(*Refresher)

// WithRegisterer is an option that allows you to register the FailedRefreshes counter to a prometheus registerer.
// If you don't register the counter, it will not be collected.
// If you register the counter more than once, it will panic.
func WithRegisterer(r prometheus.Registerer) RefresherOption {
	return func(s *Refresher) {
		r.MustRegister(s.failedRefreshes)
	}
}

// WithFailedAttemptsHook is an option that allows you to set a hook that will be called when
// the number of failed attempts exceeds the max count (WithFailedAttemptsMaxCount) or the max duration (WithFailedAttemptsMaxDuration).
func WithFailedAttemptsHook(fn func()) RefresherOption {
	return func(s *Refresher) {
		s.failedAttemptsFn = fn
	}
}

// WithFailedAttemptsMaxDuration is an option that allows you to set the max duration for failed attempts until the hook (WithFailedAttemptsHook) is called.
func WithFailedAttemptsMaxDuration(d time.Duration) RefresherOption {
	return func(s *Refresher) {
		s.failedAttemptsMaxDuration = d
	}
}

// WithFailedAttemptsMaxCount is an option that allows you to set the max count for failed attempts until the hook (WithFailedAttemptsHook) is called.
func WithFailedAttemptsMaxCount(count int32) RefresherOption {
	return func(s *Refresher) {
		s.failedAttemptsMaxCount = count
	}
}

func (r Refresher) RefreshCache(ctx context.Context, d time.Duration) {
	var failedAttempts int32 = 0
	var failedAttemptsDuration int64 = 0

	type Pair struct {
		Key   string
		Value bool
	}
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g, gCtx := errgroup.WithContext(ctx)
			g.SetLimit(r.fetchMaxConcurrency)
			resultC := make(chan Pair)

			for k, _ := range r.flags.values {
				g.Go(func() error {
					startTime := time.Now()
					r.log.Debug("refreshing flag", zap.String("key", k))
					isEnabled, err := r.fetchFn(gCtx, k)
					r.log.Debug("new value for flag", zap.String("key", k), zap.Bool("value", isEnabled))
					if err != nil {
						r.failedRefreshes.WithLabelValues(k).Inc()
						atomic.AddInt32(&failedAttempts, 1)
						atomic.AddInt64(&failedAttemptsDuration, time.Since(startTime).Nanoseconds())
						return err
					}
					resultC <- Pair{Key: k, Value: isEnabled}
					return nil
				})
			}

			go func() {
				if err := g.Wait(); err != nil {
					if r.failedAttemptsFn != nil && (failedAttempts >= r.failedAttemptsMaxCount || time.Duration(failedAttemptsDuration) > r.failedAttemptsMaxDuration) {
						r.failedAttemptsFn()
					}
					r.log.Debug("error fetching flag", zap.Error(err))
				}
				close(resultC)
			}()

			// auxMap is used to avoid locking the flags map while updating it.
			auxMap := map[string]bool{}
			for pair := range resultC {
				auxMap[pair.Key] = pair.Value
			}

			r.flags.BatchUpdate(auxMap)

			failedAttempts = 0
			failedAttemptsDuration = 0

		case <-ctx.Done():
			return
		}

	}
}
