package main

import (
	"fmt"
	"sync"
	"time"
)

type Worker struct {
	id     int
	stopCh chan struct{}
}

func (w *Worker) Start(jobsCh <-chan string, wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		for {
			select {
			case job, ok := <-jobsCh:
				if !ok {
					fmt.Printf("Worker %d: jobs channel closed, exiting\n", w.id)
					return
				}
				fmt.Printf("Worker %d processing: %s\n", w.id, job)
			case <-w.stopCh:
				fmt.Printf("Worker %d: received stop signal\n", w.id)
				return
			}
		}
	}()
}

type WorkerPool struct {
	mu      sync.Mutex
	workers map[int]*Worker
	jobsCh  chan string
	wg      sync.WaitGroup
	nextID  int
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		jobsCh:  make(chan string, 100),
		workers: make(map[int]*Worker),
	}
}

func (wp *WorkerPool) AddWorker() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	wp.nextID++
	worker := &Worker{
		id:     wp.nextID,
		stopCh: make(chan struct{}),
	}
	wp.wg.Add(1)
	worker.Start(wp.jobsCh, &wp.wg)
	wp.workers[wp.nextID] = worker
	fmt.Printf("Added worker %d\n", worker.id)
}

func (wp *WorkerPool) RemoveWorker(id int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	if len(wp.workers) == 0 {
		return
	}

	worker, ok := wp.workers[id]
	if !ok {
		fmt.Println("There is not a worker with id", id)
	}

	close(worker.stopCh)
	delete(wp.workers, id)
	fmt.Printf("Removed worker %d\n", worker.id)
}

func (wp *WorkerPool) Submit(job string) {
	wp.jobsCh <- job
}

func (wp *WorkerPool) Shutdown() {
	close(wp.jobsCh)
	wp.wg.Wait()
	fmt.Println("All workers stopped.")
}

func main() {
	pool := NewWorkerPool()

	// Добавим 3 воркера.
	for i := 0; i < 3; i++ {
		pool.AddWorker()
	}

	// Отправим несколько задач.
	for i := 0; i < 10; i++ {
		pool.Submit(fmt.Sprintf("Task #%d", i))
	}

	// Динамически добавим и удалим воркеров.
	time.Sleep(time.Second)
	pool.AddWorker()
	time.Sleep(time.Second)
	pool.RemoveWorker(3)

	// Отправим ещё задачи.
	for i := 10; i < 15; i++ {
		pool.Submit(fmt.Sprintf("Task #%d", i))
	}

	// Завершаем работу пула.
	time.Sleep(2 * time.Second)
	pool.Shutdown()
}
