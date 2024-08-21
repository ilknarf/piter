package slices

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPFlatMap(t *testing.T) {
	t.Run("Completes on no error", func(t *testing.T) {
		// order is not guaranteed
		items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		expected := []int{1, 3, 5, 7, 9}

		batchIter := slices.Chunk(items, 4)

		result, err := PFlatMap(context.Background(), batchIter, 10, func(ctx context.Context, batch []int) ([]int, error) {
			o := make([]int, 0)
			for _, n := range batch {
				if n%2 == 1 {
					o = append(o, n)
				}
			}

			return o, nil
		})

		require.NoError(t, err)
		// don't care about order
		require.ElementsMatch(t, expected, result)
	})

	t.Run("Terminates early on function error", func(t *testing.T) {
		// error on 3
		items := []int{1, 2, 3, 4, 5, 6, 7}
		failOn := 3

		funcCalls := 0

		// no concurrency so we can count when the iteration stopped
		result, err := PFlatMap(context.Background(), slices.Values(items), 0, func(ctx context.Context, i int) ([]int, error) {
			funcCalls += 1

			if i == failOn {
				return nil, errors.New("some error")
			}

			return []int{i + 1}, nil
		})

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("When Setlimit is set to 1, no concurrent executions occur", func(t *testing.T) {
		items := slices.Repeat([]int{1, 2, 3, 4}, 100000)
		it := slices.Values(items)
		mtx := &sync.Mutex{}

		res, err := PFlatMap(context.Background(), it, 1, func(ctx context.Context, i int) ([]int, error) {
			if !mtx.TryLock() {
				return nil, errors.New("unable to acquire lock")
			}

			mtx.Unlock()

			return []int{i}, nil
		})

		require.NoError(t, err)
		// we do care about order to make sure that function calls are non-concurrent
		require.Equal(t, items, res)
	})
}

func TestPMap(t *testing.T) {
	t.Run("Completes on no error", func(t *testing.T) {
		// order is not guaranteed
		items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		expected := []string{"1:2:3:4", "5:6:7:8", "9:10"}

		batchIter := slices.Chunk(items, 4)

		result, err := PMap(context.Background(), batchIter, 10, func(ctx context.Context, batch []int) (string, error) {
			nums := make([]string, 0, len(batch))
			for _, n := range batch {
				nums = append(nums, strconv.Itoa(n))
			}

			s := strings.Join(nums, ":")

			return s, nil
		})

		require.NoError(t, err)
		// don't care about order
		require.ElementsMatch(t, expected, result)
	})

	t.Run("Terminates early on function error", func(t *testing.T) {
		// error on 3
		items := []int{1, 2, 3, 4, 5, 6, 7}
		failOn := 3

		funcCalls := 0

		// no concurrency so we can count when the iteration stopped
		result, err := PMap(context.Background(), slices.Values(items), 0, func(ctx context.Context, i int) (int, error) {
			funcCalls += 1

			if i == failOn {
				return 0, errors.New("some error")
			}

			return i + 1, nil
		})

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("When Setlimit is set to 1, no concurrent executions occur", func(t *testing.T) {
		items := slices.Repeat([]int{1, 2, 3, 4}, 100000)
		it := slices.Values(items)
		mtx := &sync.Mutex{}

		res, err := PMap(context.Background(), it, 1, func(ctx context.Context, i int) (int, error) {
			if !mtx.TryLock() {
				return 0, errors.New("unable to acquire lock")
			}

			mtx.Unlock()

			return i, nil
		})

		require.NoError(t, err)
		// we do care about order to make sure that function calls are non-concurrent
		require.Equal(t, items, res)
	})
}

func TestPForEach(t *testing.T) {
	t.Run("Completes on no error", func(t *testing.T) {
		// order is not guaranteed
		items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		batchIter := slices.Chunk(items, 4)

		err := PForEach(context.Background(), batchIter, 10, func(ctx context.Context, batch []int) error {
			nums := make([]string, 0, len(batch))
			for _, n := range batch {
				nums = append(nums, strconv.Itoa(n))
			}

			_ = strings.Join(nums, ":")

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Terminates early on function error", func(t *testing.T) {
		// error on 3
		items := []int{1, 2, 3, 4, 5, 6, 7}
		failOn := 3

		funcCalls := 0

		// no concurrency so we can count when the iteration stopped
		err := PForEach(context.Background(), slices.Values(items), 1, func(ctx context.Context, i int) error {
			funcCalls += 1

			if i == failOn {
				return errors.New("some error")
			}

			return nil
		})

		require.Error(t, err)
	})

	t.Run("When Setlimit is set to 0, no concurrent executions occur", func(t *testing.T) {
		items := slices.Repeat([]int{1, 2, 3, 4}, 100000)
		it := slices.Values(items)
		mtx := &sync.Mutex{}

		err := PForEach(context.Background(), it, 0, func(ctx context.Context, i int) error {
			if !mtx.TryLock() {
				return errors.New("unable to acquire lock")
			}

			mtx.Unlock()

			return nil
		})

		require.NoError(t, err)
	})
}
