package repository

import (
	"context"

	"github.com/VaneZ444/chat-service/internal/entity"
)

type MessageRepo interface {
	CreateMessage(ctx context.Context, msg *entity.Message) error
	GetLastMessages(ctx context.Context, limit int) ([]entity.Message, error)
}
