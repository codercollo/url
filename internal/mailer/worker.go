package mailer

import "log"

//Job holds everything needed to one email
type Job struct {
	To           string
	Subject      string
	TemplateName string
	Data         any
}

//WorkerPool fans out email jobs across N goroutines
type WorkerPool struct {
	queue  chan Job
	mailer *Mailer
}

//NewworkerPool creates the pool and starts n workers
func NewWorkerPool(mailer *Mailer, workers int, bufferSize int) *WorkerPool {
	wp := &WorkerPool{
		queue:  make(chan Job, bufferSize),
		mailer: mailer,
	}
	for i := range workers {
		go wp.work(i + 1)
	}
	return wp
}

//Enqueue adds a job to the queue without blocking the caller
func (wp *WorkerPool) Enqueue(job Job) {
	select {
	case wp.queue <- job:
	default:
		log.Printf("mailer worker pool: queue full, dropping email to %s", job.To)
	}
}

//work is the goroutine loop, it pulls jobs untill the channel closes
func (wp *WorkerPool) work(id int) {
	for job := range wp.queue {
		log.Printf("mailer worker %d: processing email to %s", id, job.To)
		if err := wp.mailer.Send(
			job.To,
			job.Subject,
			job.TemplateName,
			job.Data); err != nil {
			log.Printf("mailer worker %d error sending to %s: %v", id, job.To, err)
		}
	}
}

//Shutdown drains the queue and stops workers cleanly
func (wp *WorkerPool) Shutdown() {
	close(wp.queue)
}
