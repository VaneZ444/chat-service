package usecase

import (
	"context"
	"log"
	"time"

	"github.com/VaneZ444/chat-service/internal/entity"
	"github.com/VaneZ444/chat-service/internal/repository"
)

type ChatUseCase interface {
	// теперь принимаем authorName
	SendMessage(ctx context.Context, userID int64, authorName, content string) error
	GetHistory(ctx context.Context, limit int) ([]entity.Message, error)
	SubscribeToMessages() (<-chan entity.Message, func())
}

type chatUseCase struct {
	repo      repository.MessageRepo
	broadcast chan entity.Message
}

func NewChatUseCase(repo repository.MessageRepo) ChatUseCase {
	return &chatUseCase{
		repo:      repo,
		broadcast: make(chan entity.Message, 100),
	}
}

func (uc *chatUseCase) SendMessage(ctx context.Context, userID int64, authorName, content string) error {
	msg := entity.Message{
		AuthorID:       userID,
		AuthorNickname: authorName, // сохраняем имя автора
		Content:        content,
		CreatedAt:      time.Now(),
	}

	if err := uc.repo.CreateMessage(ctx, &msg); err != nil {
		return err
	}

	uc.broadcast <- msg
	return nil
}

func (uc *chatUseCase) GetHistory(ctx context.Context, limit int) ([]entity.Message, error) {
	return uc.repo.GetLastMessages(ctx, limit)
}

func (uc *chatUseCase) SubscribeToMessages() (<-chan entity.Message, func()) {
	ch := make(chan entity.Message, 100)
	go func() {
		for msg := range uc.broadcast {
			select {
			case ch <- msg:
			default:
				log.Println("Subscriber channel is full")
			}
		}
	}()
	return ch, func() { close(ch) }
}
