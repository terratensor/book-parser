package workerpool

import (
	"fmt"
	"sync"
)

// Worker контролирует всю работу
type Worker struct {
	ID       int
	taskChan chan *Task
}

// NewWorker возвращает новый экземпляр worker-а
func NewWorker(channel chan *Task, ID int) *Worker {
	return &Worker{
		ID:       ID,
		taskChan: channel,
	}
}

// запуск worker
func (wr *Worker) Start(wg *sync.WaitGroup) {
	fmt.Printf("Starting worker %d\n", wr.ID)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for task := range wr.taskChan {
			process(wr.ID, task)
		}
	}()
}
