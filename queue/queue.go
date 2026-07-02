package queue

import (
	"errors"
	"sync"
)

type Task = func() error

type Queue struct {
	tasks            []Task
	concurrencyLimit int
	mutex            sync.Mutex
	err              error
	errMutex         sync.RWMutex
}

func New(concurrencyLimit int) (*Queue, error) {
	if concurrencyLimit < 1 {
		return nil, errors.New("concurrency limit must be greater than zero")
	}
	return &Queue{
		tasks:            make([]Task, 0),
		concurrencyLimit: concurrencyLimit,
	}, nil
}

func (q *Queue) next() (Task, bool) {
	var zero Task
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if len(q.tasks) == 0 {
		return zero, false
	}
	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	return task, true
}

func (q *Queue) hasError() bool {
	q.errMutex.RLock()
	defer q.errMutex.RUnlock()
	return q.err != nil
}

func (q *Queue) worker(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		if q.hasError() {
			return
		}

		task, ok := q.next()
		if !ok {
			return
		}
		err := task()
		if err != nil {
			q.errMutex.Lock()
			q.err = err
			q.errMutex.Unlock()
			return
		}
	}
}

func (q *Queue) Enqueue(task Task) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.tasks = append(q.tasks, task)
}

func (q *Queue) ProcessAll() error {
	var wg sync.WaitGroup
	wg.Add(q.concurrencyLimit)

	for i := 0; i < q.concurrencyLimit; i++ {
		go q.worker(&wg)
	}
	wg.Wait()

	return q.err
}
