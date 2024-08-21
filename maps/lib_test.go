package maps

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPMap(t *testing.T) {
	t.Run("Completes on no error", func(t *testing.T) {
		// order is not guaranteed
		items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		expected := make(map[string]string)
		for k, v := range items {
			expected[strconv.Itoa(v)] = strconv.Itoa(k)
		}

		iter := slices.All(items)

		result, err := PMap(context.Background(), iter, 10, func(ctx context.Context, k int, v int) (string, string, error) {
			time.Sleep(50 * time.Millisecond)
			return strconv.Itoa(v), strconv.Itoa(k), nil
		})

		require.NoError(t, err)
		// don't care about order
		require.Equal(t, expected, result)
	})

	t.Run("Terminates on function error", func(t *testing.T) {
		// error on 3
		items := []string{"1", "2", "3", "4", "5", "6", "7"}
		failOn := "3"

		// no concurrency so we can count when the iteration stopped
		result, err := PMap(context.Background(), slices.All(items), 1, func(ctx context.Context, k int, v string) (string, string, error) {

			if v == failOn {
				return "", "", errors.New("some error")
			}

			return v, v, nil
		})

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("When Setlimit is set to 1, no concurrent executions occur", func(t *testing.T) {
		items := slices.Repeat([]int{1, 2, 3, 4}, 10)
		it := slices.All(items)
		mtx := &sync.Mutex{}

		expected := maps.Collect(slices.All(items))

		res, err := PMap(context.Background(), it, 1, func(ctx context.Context, k int, v int) (int, int, error) {
			if !mtx.TryLock() {
				return 0, 0, errors.New("unable to acquire lock")
			}

			mtx.Unlock()

			return k, v, nil
		})

		require.NoError(t, err)
		// we do care about order to make sure that function calls are non-concurrent
		require.Equal(t, expected, res)
	})
}

func TestPForEach(t *testing.T) {
	t.Run("Completes on no error", func(t *testing.T) {
		// order is not guaranteed
		items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		iter := slices.All(items)

		err := PForEach(context.Background(), iter, 10, func(ctx context.Context, k int, v int) error {

			_ = fmt.Sprintf("%d:%d", k, v)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Terminates early on function error", func(t *testing.T) {
		// error on 3
		items := []int{1, 2, 3, 4, 5, 6, 7}
		failOn := 3

		funcCalls := 0
		var errOnCall int

		// no concurrency so we can count when the iteration stopped
		err := PForEach(context.Background(), slices.All(items), 1, func(ctx context.Context, k int, v int) error {
			funcCalls += 1

			if v == failOn {
				errOnCall = funcCalls
				return errors.New("some error")
			}

			return nil
		})

		require.Error(t, err)
		require.Equal(t, errOnCall, funcCalls)
	})

	t.Run("When Setlimit is set to 1, no concurrent executions occur", func(t *testing.T) {
		items := slices.Repeat([]string{"asdf", "12123144", "awgwegwe", "1231241"}, 50)
		it := slices.All(items)
		mtx := &sync.Mutex{}

		err := PForEach(context.Background(), it, 1, func(ctx context.Context, k int, v string) error {
			if !mtx.TryLock() {
				return errors.New("unable to acquire lock")
			}

			mtx.Unlock()

			return nil
		})

		require.NoError(t, err)
	})
}
