package usecase

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/VaneZ444/chat-service/internal/entity"
	"github.com/VaneZ444/chat-service/internal/repository"
)

// Интерфейс
type ChatUseCase interface {
	SendMessage(ctx context.Context, userID int64, authorName, content string) error
	GetHistory(ctx context.Context, limit int) ([]entity.Message, error)
	SubscribeToMessages() (<-chan entity.Message, func())
}

// Реализация
type chatUseCase struct {
	repo        repository.MessageRepo
	subscribers map[chan entity.Message]struct{}
	mu          sync.RWMutex
}

func NewChatUseCase(repo repository.MessageRepo) ChatUseCase {
	return &chatUseCase{
		repo:        repo,
		subscribers: make(map[chan entity.Message]struct{}),
	}
}

func (uc *chatUseCase) SendMessage(ctx context.Context, userID int64, authorName, content string) error {
	msg := entity.Message{
		AuthorID:       userID,
		AuthorNickname: authorName,
		Content:        content,
		CreatedAt:      time.Now(),
	}

	if err := uc.repo.CreateMessage(ctx, &msg); err != nil {
		return err
	}

	uc.mu.RLock()
	defer uc.mu.RUnlock()
	for ch := range uc.subscribers {
		select {
		case ch <- msg:
		default:
			log.Println("subscriber channel full")
		}
	}
	return nil
}

func (uc *chatUseCase) GetHistory(ctx context.Context, limit int) ([]entity.Message, error) {
	return uc.repo.GetLastMessages(ctx, limit)
}

func (uc *chatUseCase) SubscribeToMessages() (<-chan entity.Message, func()) {
	ch := make(chan entity.Message, 100)

	uc.mu.Lock()
	uc.subscribers[ch] = struct{}{}
	uc.mu.Unlock()

	unsubscribe := func() {
		uc.mu.Lock()
		delete(uc.subscribers, ch)
		close(ch)
		uc.mu.Unlock()
	}
	return ch, unsubscribe
}
