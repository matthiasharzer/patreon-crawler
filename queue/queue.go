package queue

import "sync"

type Item[T any] struct {
	value   T
	process func(item T) error
}

type Queue[T any] struct {
	items            []Item[T]
	concurrencyLimit int
	mutex            sync.Mutex
	err              error
	errMutex         sync.RWMutex
}

func New[T any](concurrencyLimit int) *Queue[T] {
	return &Queue[T]{
		items:            make([]Item[T], 0),
		concurrencyLimit: concurrencyLimit,
	}
}

func (q *Queue[T]) next() (Item[T], bool) {
	var zero Item[T]
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if len(q.items) == 0 {
		return zero, false
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}

func (q *Queue[T]) hasError() bool {
	q.errMutex.RLock()
	defer q.errMutex.RUnlock()
	return q.err != nil
}

func (q *Queue[T]) worker(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		if q.hasError() {
			return
		}

		item, ok := q.next()
		if !ok {
			return
		}
		err := item.process(item.value)
		if err != nil {
			q.errMutex.Lock()
			q.err = err
			q.errMutex.Unlock()
			return
		}
	}
}

func (q *Queue[T]) Enqueue(item T, process func(item T) error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, Item[T]{value: item, process: process})
}

func (q *Queue[T]) ProcessAll() error {
	var wg sync.WaitGroup
	wg.Add(q.concurrencyLimit)

	for i := 0; i < q.concurrencyLimit; i++ {
		go q.worker(&wg)
	}
	wg.Wait()

	return q.err
}
