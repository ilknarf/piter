package maps

import (
	"context"
	"iter"
	"sync"

	"golang.org/x/sync/errgroup"
)

// PForEach performs func f in a goroutine pool of t threads, exiting early if there is an error
func PForEach[I1, I2 any](ctx context.Context, seq iter.Seq2[I1, I2], threads int, f func(context.Context, I1, I2) error) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(threads)

	seq(func(v1 I1, v2 I2) bool {
		// stop processing if error is encountered in errgroup
		if err := ctx.Err(); err != nil {
			return false
		}

		eg.Go(func() error {
			if err := ctx.Err(); err != nil {
				return err
			}

			return f(ctx, v1, v2)
		})

		return true
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

// PMap just wraps PForEach in a closure to add keys. If two keys are the same (e.g. Seq2 is not from a map), there is no way of determining which will be present in the resulting map
func PMap[I1, I2 any, K comparable, V any](outerCtx context.Context, seq iter.Seq2[I1, I2], threads int, f func(context.Context, I1, I2) (K, V, error)) (map[K]V, error) {
	res := make(map[K]V)
	mtx := &sync.Mutex{}

	err := PForEach(outerCtx, seq, threads, func(innerCtx context.Context, v1 I1, v2 I2) error {
		k, v, err := f(innerCtx, v1, v2)
		if err != nil {
			return err
		}

		mtx.Lock()
		defer mtx.Unlock()
		res[k] = v

		return nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}
