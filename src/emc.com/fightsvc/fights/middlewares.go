package fights

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   Service
	logger log.Logger
}

func (mw loggingMiddleware) PostFight(ctx context.Context, p Fight) (fight map[string]interface{}, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "PostFight", "id", p.ID, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.PostFight(ctx, p)
}

func (mw loggingMiddleware) GetFight(ctx context.Context, id string) (fight map[string]interface{}, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GetFight", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.GetFight(ctx, id)
}

func (mw loggingMiddleware) PutFight(ctx context.Context, id string, attack map[string]interface{}) (fight map[string]interface{}, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "PutFight", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.PutFight(ctx, id, attack)
}

func (mw loggingMiddleware) DeleteFight(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "DeleteFight", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.DeleteFight(ctx, id)
}
