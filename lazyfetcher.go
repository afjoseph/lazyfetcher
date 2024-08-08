package lazyfetcher

import (
	"context"
	"sync"
	"time"

	"github.com/go-playground/errors/v5"
)

type decayablePriorityList[T any] struct {
	values             []T
	highestPriorityIdx int
	lastFetchedAt      time.Time
}

// LazyFetcher allows inserting and fetching a map[string]T
// It has one interface: Fetch(). When Fetch() is called, one of three things happens:
//  1. if the key does not exist in the map, calls fetcher(key) and stores the
//     result in the map, then return it
//  2. If the key exists, but has expired, calls fetcher(key) and stores the
//     result in the map, then return it and refresh the expiration time
//  3. If the key exists and has not expired, return the value in the map
type LazyFetcher[T any] struct {
	fetcher func(context.Context, string) ([]T, int, error)
	entries sync.Map
	// entries    map[string]*entry[T]
	decayEvery time.Duration
}

func New[T any](
	decayEvery time.Duration,
	fetcher func(context.Context, string) ([]T, int, error),
) *LazyFetcher[T] {
	return &LazyFetcher[T]{
		fetcher:    fetcher,
		entries:    sync.Map{},
		decayEvery: decayEvery,
	}
}

func (lf *LazyFetcher[T]) FetchPriority(
	ctx context.Context,
	key string,
) (T, error) {
	ls, activeIdx, err := lf.Fetch(ctx, key)
	if err != nil {
		// Because Go doesn't have a nullable constraint
		// https://github.com/golang/go/issues/53656
		var null T
		return null, err
	}
	return ls[activeIdx], nil
}

func (lf *LazyFetcher[T]) Fetch(
	ctx context.Context,
	key string,
) ([]T, int, error) {
	// Check if the key exists in the map
	if entry, ok := lf.entries.Load(key); ok {
		typedEntry := entry.(*decayablePriorityList[T])
		// Check if the entry has expired
		if time.Since(typedEntry.lastFetchedAt) > lf.decayEvery {
			// If the entry has expired, fetch a new value and update the entry
			newValues, newHighestPriorityIdx, err := lf.fetcher(ctx, key)
			if err != nil {
				return nil, typedEntry.highestPriorityIdx, errors.Wrapf(
					err,
					"fetching %s",
					key,
				)
			}
			typedEntry.values = newValues
			typedEntry.highestPriorityIdx = newHighestPriorityIdx
			typedEntry.lastFetchedAt = time.Now()
			lf.entries.Store(key, typedEntry)
			return newValues, newHighestPriorityIdx, nil
		}
		// If the entry has not expired, return the value
		return typedEntry.values, typedEntry.highestPriorityIdx, nil
	}

	// If the key does not exist in the map, fetch a new value and add it to the map
	newValues, newHighestPriorityIdx, err := lf.fetcher(ctx, key)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "fetching %s", key)
	}
	lf.entries.Store(key, &decayablePriorityList[T]{
		values:             newValues,
		highestPriorityIdx: newHighestPriorityIdx,
		lastFetchedAt:      time.Now(),
	})
	return newValues, newHighestPriorityIdx, nil
}
