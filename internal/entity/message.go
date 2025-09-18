package entity

import (
	"time"
)

type Message struct {
	ID             int64     `json:"id"`
	AuthorID       int64     `json:"author_id"`
	AuthorNickname string    `json:"author_nickname"` // новое поле
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}
