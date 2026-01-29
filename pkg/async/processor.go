package async

import (
	"context"
	"errors"
	"sync"
	"time"

	"wechat-service/internal/config"
	"wechat-service/pkg/logger"
)

// Task represents an async task
type Task struct {
	ID        string
	Type      string
	Payload   interface{}
	Retry     int
	MaxRetry  int
	CreatedAt time.Time
	Execute   func(ctx context.Context, payload interface{}) error
}

// Processor handles async task processing
type Processor struct {
	cfg      *config.Config
	log      *logger.Logger
	queue    chan *Task
	workers  int
	wg       sync.WaitGroup
	stopCh   chan struct{}
	mu       sync.RWMutex
	stats    ProcessorStats
}

// ProcessorStats holds processing statistics
type ProcessorStats struct {
	TotalProcessed int64
	TotalFailed    int64
	TotalRetried   int64
	QueueSize      int
}

// NewProcessor creates a new async processor
func NewProcessor(cfg *config.Config, log *logger.Logger) *Processor {
	p := &Processor{
		cfg:     cfg,
		log:     log,
		queue:   make(chan *Task, cfg.Async.QueueSize),
		workers: cfg.Async.Workers,
		stopCh:  make(chan struct{}),
	}

	// Start workers
	for i := 0; i < cfg.Async.Workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	return p
}

// worker is a background worker
func (p *Processor) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case task := <-p.queue:
			p.processTask(task)
		case <-p.stopCh:
			p.log.Debug("Worker stopping", "id", id)
			return
		}
	}
}

// processTask processes a single task
func (p *Processor) processTask(task *Task) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := task.Execute(ctx, task.Payload)
	if err != nil {
		p.log.Error("Task failed",
			"task_id", task.ID,
			"task_type", task.Type,
			"retry", task.Retry,
			"error", err,
		)

		p.mu.Lock()
		p.stats.TotalFailed++
		p.mu.Unlock()

		// Retry if under limit
		if task.Retry < task.MaxRetry {
			task.Retry++
			p.mu.Lock()
			p.stats.TotalRetried++
			p.mu.Unlock()

			// Requeue with delay
			go func() {
				time.Sleep(time.Duration(p.cfg.Async.RetryDelay) * time.Millisecond)
				p.queue <- task
			}()
		} else {
			p.log.Error("Task exceeded max retries, dropping",
				"task_id", task.ID,
				"task_type", task.Type,
				"max_retry", task.MaxRetry,
			)
		}
		return
	}

	p.mu.Lock()
	p.stats.TotalProcessed++
	p.mu.Unlock()
}

// Submit submits a task for async processing
func (p *Processor) Submit(task *Task) error {
	select {
	case p.queue <- task:
		p.mu.Lock()
		p.stats.QueueSize = len(p.queue)
		p.mu.Unlock()
		return nil
	default:
		p.log.Warn("Async queue full, dropping task",
			"task_id", task.ID,
			"queue_size", len(p.queue),
		)
		return errors.New("queue full")
	}
}

// SubmitFunc submits a task with execute function
func (p *Processor) SubmitFunc(taskType string, payload interface{}, execute func(ctx context.Context, payload interface{}) error) error {
	task := &Task{
		ID:        generateID(),
		Type:      taskType,
		Payload:   payload,
		Execute:   execute,
		MaxRetry:  p.cfg.Async.RetryCount,
		CreatedAt: time.Now(),
	}
	return p.Submit(task)
}

// SubmitWithRetry submits with custom retry count
func (p *Processor) SubmitWithRetry(taskType string, payload interface{}, execute func(ctx context.Context, payload interface{}) error, maxRetry int) error {
	task := &Task{
		ID:        generateID(),
		Type:      taskType,
		Payload:   payload,
		Execute:   execute,
		MaxRetry:  maxRetry,
		CreatedAt: time.Now(),
	}
	return p.Submit(task)
}

// GetStats returns current statistics
func (p *Processor) GetStats() ProcessorStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return ProcessorStats{
		TotalProcessed: p.stats.TotalProcessed,
		TotalFailed:    p.stats.TotalFailed,
		TotalRetried:   p.stats.TotalRetried,
		QueueSize:      len(p.queue),
	}
}

// Stop stops the processor
func (p *Processor) Stop() {
	close(p.stopCh)
	p.wg.Wait()
	p.log.Info("Async processor stopped", "stats", p.GetStats())
}

// generateID generates a unique task ID
func generateID() string {
	return time.Now().Format("20060102150405") + randomString(8)
}

// randomString generates a random string
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
