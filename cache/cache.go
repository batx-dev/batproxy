package cache

import (
	"context"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
)

const (
	CacheKeyAll = "_all_"

	DefaultExpiration     = 15 * time.Second
	DefaultExitExpiration = time.Hour
	DefaultCapacity       = 10_000
)

type ListFunc[I comparable, V any] func(ctx context.Context, ids []I) (results []*Result[I, V], err error)

type Result[I comparable, V any] struct {
	ID       I
	Val      *V
	ExitTime time.Time
}

func listFunc[I comparable, V any](
	cacheResults *cache.Cache[I, *V],
	keyAll I,
	expiration,
	exitExpiration time.Duration,
	nextList ListFunc[I, V],
) ListFunc[I, V] {
	return func(ctx context.Context, ids []I) (results []*Result[I, V], err error) {
		// List all vals
		if len(ids) == 0 {
			// Try to list vals from cache
			if _, ok := cacheResults.Get(keyAll); ok {
				for _, id := range cacheResults.Keys() {
					if id == keyAll {
						continue
					}
					if val, ok := cacheResults.Get(id); ok && val != nil {
						results = append(results, &Result[I, V]{
							ID:  id,
							Val: val,
						})
					}
				}
				return results, nil
			}

			// Then list vals from backend service if not in cache.
			nextResults, err := nextList(ctx, ids)
			if err != nil {
				return nil, err
			}
			results = append(results, nextResults...)

			// Update vals to cache.
			for _, result := range nextResults {
				if result.ExitTime.IsZero() {
					cacheResults.Set(result.ID, result.Val, cache.WithExpiration(expiration))
				} else {
					cacheResults.Set(result.ID, result.Val, cache.WithExpiration(exitExpiration))
				}
			}
			cacheResults.Set(
				keyAll,
				nil,
				cache.WithExpiration(expiration-time.Second),
			)

			return results, nil
		}

		// Try to list vals from cache.
		var notFoundInCacheIDs []I
		for _, id := range ids {
			val, ok := cacheResults.Get(id)
			if !ok {
				notFoundInCacheIDs = append(notFoundInCacheIDs, id)
			} else {
				if val != nil {
					results = append(results, &Result[I, V]{
						ID:  id,
						Val: val,
					})
				}
				// val == nil means cache not found val, so just ignore it.
			}
		}
		if len(notFoundInCacheIDs) == 0 {
			return results, nil
		}

		// Then list vals from backend service if not found in cache.
		nextResults, err := nextList(ctx, notFoundInCacheIDs)
		if err != nil {
			return nil, err
		}
		results = append(results, nextResults...)

		// Update vals to cache.
		nextResultsMap := make(map[I]*Result[I, V])
		for _, result := range nextResults {
			nextResultsMap[result.ID] = result
		}
		for _, id := range notFoundInCacheIDs {
			result, ok := nextResultsMap[id]
			if !ok {
				cacheResults.Set(id, nil, cache.WithExpiration(expiration))
			} else {
				if result.ExitTime.IsZero() {
					cacheResults.Set(id, result.Val, cache.WithExpiration(expiration))
				} else {
					cacheResults.Set(id, result.Val, cache.WithExpiration(exitExpiration))
				}
			}
		}

		return results, nil
	}
}
