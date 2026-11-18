package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/metrics"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

var ErrQueueFull = errors.New("telegram update queue is full")

type Runner struct {
	router         *Router
	dedup          ports.DedupStore
	logger         *logger.Logger
	queue          chan models.Update
	workers        int
	dedupTTL       int
	processTimeout time.Duration
	chatLocks      *keyedMutex
	mu             sync.RWMutex
	closed         bool
	startOnce      sync.Once
	stopOnce       sync.Once
	wg             sync.WaitGroup
}

func NewRunner(
	router *Router,
	dedup ports.DedupStore,
	logger *logger.Logger,
	workers int,
	queueSize int,
	dedupTTL time.Duration,
	processTimeout time.Duration,
) *Runner {
	if workers < 1 {
		workers = 1
	}
	if queueSize < workers {
		queueSize = workers * 8
	}
	return &Runner{
		router:         router,
		dedup:          dedup,
		logger:         logger,
		queue:          make(chan models.Update, queueSize),
		workers:        workers,
		dedupTTL:       int(dedupTTL.Seconds()),
		processTimeout: processTimeout,
		chatLocks:      newKeyedMutex(),
	}
}

func (r *Runner) Start(ctx context.Context) {
	r.startOnce.Do(func() {
		for index := 0; index < r.workers; index++ {
			r.wg.Add(1)
			go r.worker(ctx)
		}
		r.logger.Info("telegram runner started", zap.Int("workers", r.workers), zap.Int("queue", cap(r.queue)))
	})
}

func (r *Runner) SubmitRaw(ctx context.Context, raw []byte) error {
	var update models.Update
	if err := json.Unmarshal(raw, &update); err != nil {
		return fmt.Errorf("unmarshal telegram update: %w", err)
	}
	if update.Message == nil && update.CallbackQuery == nil {
		return nil
	}

	key := fmt.Sprintf("telegram:update:%d", update.ID)
	if update.ID != 0 {
		claimed, err := r.dedup.Claim(ctx, key, r.dedupTTL)
		if err != nil {
			r.logger.Warn("update dedup claim failed", zap.Error(err))
		} else if !claimed {
			r.logger.Debug("duplicate telegram update ignored")
			return nil
		}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		if update.ID != 0 {
			_ = r.dedup.Release(ctx, key)
		}
		return ErrQueueFull
	}
	select {
	case r.queue <- update:
		return nil
	default:
		if update.ID != 0 {
			_ = r.dedup.Release(ctx, key)
		}
		return ErrQueueFull
	}
}

func (r *Runner) Stop(ctx context.Context) {
	r.stopOnce.Do(func() {
		r.mu.Lock()
		if r.closed {
			r.mu.Unlock()
			return
		}
		r.closed = true
		close(r.queue)
		r.mu.Unlock()

		done := make(chan struct{})
		go func() {
			r.wg.Wait()
			close(done)
		}()
		select {
		case <-done:
			r.logger.Info("telegram runner stopped")
		case <-ctx.Done():
			r.logger.Warn("telegram runner stop timed out")
		}
	})
}

func (r *Runner) worker(ctx context.Context) {
	defer r.wg.Done()
	for update := range r.queue {
		r.process(ctx, update)
	}
}

func (r *Runner) process(ctx context.Context, update models.Update) {
	started := time.Now()
	processCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.processTimeout)
	defer cancel()

	unlock := r.chatLocks.Lock(updateChatID(update))
	defer unlock()
	defer func() {
		if recovered := recover(); recovered != nil {
			metrics.RecordUpdate(updateKind(update), false, time.Since(started))
			r.logger.Error("telegram update panic", zap.Any("panic", recovered))
		}
	}()

	kind := updateKind(update)
	if err := r.router.Process(processCtx, update); err != nil {
		metrics.RecordUpdate(kind, false, time.Since(started))
		r.logger.Error("telegram update processing failed", zap.String("kind", kind), zap.Error(err))
		return
	}
	metrics.RecordUpdate(kind, true, time.Since(started))
}

func updateChatID(update models.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil {
		if update.CallbackQuery.Message.Message != nil {
			return update.CallbackQuery.Message.Message.Chat.ID
		}
		if update.CallbackQuery.Message.InaccessibleMessage != nil {
			return update.CallbackQuery.Message.InaccessibleMessage.Chat.ID
		}
		return update.CallbackQuery.From.ID
	}
	return 0
}

func updateKind(update models.Update) string {
	switch {
	case update.CallbackQuery != nil:
		return "callback_query"
	case update.Message != nil:
		return "message"
	default:
		return "other"
	}
}

type keyedMutexEntry struct {
	mu   sync.Mutex
	refs int
}

type keyedMutex struct {
	mu      sync.Mutex
	entries map[int64]*keyedMutexEntry
}

func newKeyedMutex() *keyedMutex {
	return &keyedMutex{entries: make(map[int64]*keyedMutexEntry)}
}

func (m *keyedMutex) Lock(key int64) func() {
	m.mu.Lock()
	entry, ok := m.entries[key]
	if !ok {
		entry = &keyedMutexEntry{}
		m.entries[key] = entry
	}
	entry.refs++
	m.mu.Unlock()

	entry.mu.Lock()
	return func() {
		entry.mu.Unlock()
		m.mu.Lock()
		entry.refs--
		if entry.refs == 0 {
			delete(m.entries, key)
		}
		m.mu.Unlock()
	}
}
