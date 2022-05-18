package graph

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	core "github.com/authzed/spicedb/pkg/proto/core/v1"

	"github.com/authzed/spicedb/internal/dispatch"
	v1 "github.com/authzed/spicedb/pkg/proto/dispatch/v1"
	"github.com/authzed/spicedb/pkg/tuple"
)

type ParallelChecker struct {
	toCheck       chan *v1.DispatchCheckRequest
	c             dispatch.Check
	g             *errgroup.Group
	checkCtx      context.Context
	subject       *core.ObjectAndRelation
	maxConcurrent uint8
	results       *tuple.ONRSet
	mu            sync.Mutex
}

func NewParallelChecker(ctx context.Context, c dispatch.Check, subject *core.ObjectAndRelation, maxConcurrent uint8) *ParallelChecker {
	g, checkCtx := errgroup.WithContext(ctx)
	toCheck := make(chan *v1.DispatchCheckRequest)
	return &ParallelChecker{toCheck, c, g, checkCtx, subject, maxConcurrent, tuple.NewONRSet(), sync.Mutex{}}
}

func (pc *ParallelChecker) AddResult(resource *core.ObjectAndRelation) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.results.Add(resource)
}

func (pc *ParallelChecker) QueueCheck(resource *core.ObjectAndRelation, meta *v1.ResolverMeta) {
	pc.toCheck <- &v1.DispatchCheckRequest{
		Metadata:          meta,
		ObjectAndRelation: resource,
		Subject:           pc.subject,
	}
}

func (pc *ParallelChecker) Start() {
	pc.g.Go(func() error {
		sem := semaphore.NewWeighted(int64(pc.maxConcurrent))
		for {
			if err := sem.Acquire(pc.checkCtx, 1); err != nil {
				return err
			}
			req, ok := <-pc.toCheck
			if !ok {
				sem.Release(1)
				break
			}

			pc.g.Go(func() error {
				defer sem.Release(1)
				res, err := pc.c.DispatchCheck(pc.checkCtx, req)
				if err != nil {
					return err
				}
				if res.Membership == v1.DispatchCheckResponse_MEMBER {
					pc.AddResult(req.ObjectAndRelation)
				}
				return nil
			})
		}
		if err := sem.Acquire(pc.checkCtx, int64(pc.maxConcurrent)); err != nil {
			return err
		}
		return nil
	})
}

func (pc *ParallelChecker) Finish() (*tuple.ONRSet, error) {
	close(pc.toCheck)
	if err := pc.g.Wait(); err != nil {
		return nil, err
	}

	return pc.results, nil
}
