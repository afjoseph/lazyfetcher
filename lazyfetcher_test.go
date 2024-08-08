package lazyfetcher

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-playground/errors/v5"
	"github.com/stretchr/testify/require"
)

func TestLazyFetcher(t *testing.T) {
	t.Run("New LazyFetcher", func(t *testing.T) {
		decayEvery := time.Minute
		fetcher := func(ctx context.Context, key string) ([]string, int, error) {
			return []string{"value"}, 0, nil
		}

		lf := New[string](decayEvery, fetcher)

		require.NotNil(t, lf)
		require.Equal(t, decayEvery, lf.decayEvery)
		require.NotNil(t, lf.fetcher)
		// require.Len(t, lf.entries.Len(), 0)
	})

	t.Run("Fetch non-existent key", func(t *testing.T) {
		fetcher := func(ctx context.Context, key string) ([]string, int, error) {
			return []string{"new value"}, 0, nil
		}

		lf := New[string](time.Minute, fetcher)

		values, priority, err := lf.Fetch(context.Background(), "key1")
		require.NoError(t, err)
		require.Len(t, values, 1)
		require.Contains(t, values, "new value")
		require.Equal(t, 0, priority)
	})

	t.Run("Fetch existing non-expired key", func(t *testing.T) {
		fetchCount := 0
		fetcher := func(ctx context.Context, key string) ([]string, int, error) {
			fetchCount++
			return []string{fmt.Sprintf("value%d", fetchCount)}, 0, nil
		}

		lf := New[string](time.Minute, fetcher)

		// First fetch
		values1, priority1, err := lf.Fetch(context.Background(), "key1")
		require.NoError(t, err)
		require.Len(t, values1, 1)
		require.Contains(t, values1, "value1")
		require.Equal(t, 0, priority1)

		// Second fetch (should return cached value)
		values2, priority2, err := lf.Fetch(context.Background(), "key1")
		require.NoError(t, err)
		require.Len(t, values2, 1)
		require.Contains(t, values2, "value1")
		require.Equal(t, 0, priority2)

		require.Equal(t, 1, fetchCount)
	})

	t.Run("Fetch existing expired key", func(t *testing.T) {
		fetchCount := 0
		fetcher := func(ctx context.Context, key string) ([]string, int, error) {
			fetchCount++
			return []string{fmt.Sprintf("value%d", fetchCount)}, 0, nil
		}

		lf := New[string](time.Millisecond, fetcher)

		// First fetch
		values1, priority1, err := lf.Fetch(context.Background(), "key1")
		require.NoError(t, err)
		require.Len(t, values1, 1)
		require.Contains(t, values1, "value1")
		require.Equal(t, 0, priority1)

		// Wait for expiration
		time.Sleep(2 * time.Millisecond)

		// Second fetch (should fetch new value)
		values2, priority2, err := lf.Fetch(context.Background(), "key1")
		require.NoError(t, err)
		require.Len(t, values2, 1)
		require.Contains(t, values2, "value2")
		require.Equal(t, 0, priority2)

		require.Equal(t, 2, fetchCount)
	})

	t.Run("Fetch with error", func(t *testing.T) {
		fetcher := func(ctx context.Context, key string) ([]string, int, error) {
			return nil, 0, errors.New("fetch error")
		}

		lf := New[string](time.Minute, fetcher)

		values, priority, err := lf.Fetch(context.Background(), "key1")
		require.Error(t, err)
		require.Nil(t, values)
		require.Equal(t, 0, priority)
		require.Contains(t, err.Error(), "fetch error")
	})

	t.Run("Fetch with error on expired key", func(t *testing.T) {
		fetchCount := 0
		fetcher := func(ctx context.Context, key string) ([]string, int, error) {
			fetchCount++
			if fetchCount == 1 {
				return []string{"initial value"}, 0, nil
			}
			return nil, 0, errors.New("fetch error")
		}

		lf := New[string](time.Millisecond, fetcher)

		// First fetch
		values1, priority1, err := lf.Fetch(context.Background(), "key1")
		require.NoError(t, err)
		require.Len(t, values1, 1)
		require.Contains(t, values1, "initial value")
		require.Equal(t, 0, priority1)

		// Wait for expiration
		time.Sleep(2 * time.Millisecond)

		// Second fetch (should return error but keep old value)
		values2, priority2, err := lf.Fetch(context.Background(), "key1")
		require.Error(t, err)
		require.Contains(t, err.Error(), "fetch error")
		require.Len(t, values2, 1)
		require.Contains(t, values2, "initial value")
		require.Equal(t, 0, priority2)

		require.Equal(t, 2, fetchCount)
	})

	t.Run("Multiple fetches yield non-expired value", func(t *testing.T) {
		fetchCount := 0
		fetcher := func(ctx context.Context, key string) ([]string, int, error) {
			fetchCount++
			time.Sleep(10 * time.Millisecond) // Simulate slow fetch
			return []string{fmt.Sprintf("value%d", fetchCount)}, 0, nil
		}

		lf := New[string](time.Minute, fetcher)

		for i := 0; i < 10; i++ {
			v, p, err := lf.Fetch(context.Background(), "key1")
			require.NoError(t, err)
			require.Len(t, v, 1)
			require.Contains(t, v, "value1")
			require.Equal(t, 0, p)
		}
	})
}
