package queue_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"patreon-crawler/queue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	t.Run("process all items with concurrency 1 preserves order", func(t *testing.T) {
		items := []int{1, 2, 3, 4, 5}
		processed := make([]int, 0, len(items))
		var mu sync.Mutex

		q1 := queue.New[int](1)
		for _, it := range items {
			it := it
			q1.Enqueue(it, func(v int) error {
				mu.Lock()
				processed = append(processed, v)
				mu.Unlock()
				return nil
			})
		}

		err := q1.ProcessAll()
		require.NoError(t, err)

		require.Equal(t, len(items), len(processed))
		assert.Equal(t, items, processed)
	})

	t.Run("concurrency > 1 processes each item exactly once", func(t *testing.T) {
		n := 100
		conc := 8
		counts := make(map[int]int, n)
		var mu2 sync.Mutex

		q2 := queue.New[int](conc)
		for i := 0; i < n; i++ {
			i := i
			q2.Enqueue(i, func(v int) error {
				mu2.Lock()
				counts[v]++
				mu2.Unlock()
				return nil
			})
		}

		err := q2.ProcessAll()
		require.NoError(t, err)

		require.Equal(t, n, len(counts))
		for i := 0; i < n; i++ {
			assert.Equal(t, 1, counts[i], "item %d processed count", i)
		}
	})

	t.Run("error stops further processing with concurrency 1", func(t *testing.T) {
		errorItems := []int{10, 20, 30, 40, 50}
		var processedCount int32
		failOn := 30

		q3 := queue.New[int](1)
		for _, it := range errorItems {
			it := it
			q3.Enqueue(it, func(v int) error {
				atomic.AddInt32(&processedCount, 1)
				if v == failOn {
					return errors.New("boom")
				}
				return nil
			})
		}

		err := q3.ProcessAll()
		require.Error(t, err)

		assert.Equal(t, int32(3), processedCount)
	})

	t.Run("max parallelism does not exceed concurrency limit", func(t *testing.T) {
		nParallel := 40
		concLimit := 5
		var active int32
		var maxActive int32
		var mu3 sync.Mutex

		q4 := queue.New[int](concLimit)
		for i := 0; i < nParallel; i++ {
			q4.Enqueue(i, func(v int) error {
				cur := atomic.AddInt32(&active, 1)
				mu3.Lock()
				if cur > maxActive {
					maxActive = cur
				}
				mu3.Unlock()
				time.Sleep(10 * time.Millisecond)
				atomic.AddInt32(&active, -1)
				return nil
			})
		}

		err := q4.ProcessAll()
		require.NoError(t, err)

		assert.LessOrEqual(t, int(maxActive), concLimit)
	})

	t.Run("empty queue returns nil error", func(t *testing.T) {
		q5 := queue.New[int](3)
		err := q5.ProcessAll()
		assert.NoError(t, err)
	})
}
