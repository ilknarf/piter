package slices

import (
	"context"
	"iter"
	"sync"

	"golang.org/x/sync/errgroup"
)

func PForEach[I any](ctx context.Context, seq iter.Seq[I], threads int, f func(context.Context, I) error) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(threads)

	seq(func(v I) bool {
		// stop processing if error is encountered in errgroup
		if err := ctx.Err(); err != nil {
			return false
		}

		eg.Go(func() error {
			// since blocking happens above, we want to check inside the goroutine as well
			if err := ctx.Err(); err != nil {
				return err
			}

			return f(ctx, v)
		})

		return true
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

// PFlatMap just wraps PForEach in a closure to append values
func PFlatMap[I, O any](outerCtx context.Context, seq iter.Seq[I], threads int, f func(context.Context, I) ([]O, error)) ([]O, error) {
	res := make([]O, 0)
	mtx := &sync.Mutex{}

	err := PForEach(outerCtx, seq, threads, func(innerCtx context.Context, v I) error {
		o, err := f(innerCtx, v)
		if err != nil {
			return err
		}

		mtx.Lock()
		defer mtx.Unlock()
		res = append(res, o...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

// PMap just wraps PFlatMap and returns a single-element list on success
func PMap[I, O any](outerCtx context.Context, seq iter.Seq[I], threads int, f func(context.Context, I) (O, error)) ([]O, error) {
	return PFlatMap(outerCtx, seq, threads, func(ctx context.Context, i I) ([]O, error) {
		o, err := f(ctx, i)
		if err != nil {
			return nil, err
		}

		return []O{o}, nil
	})
}
