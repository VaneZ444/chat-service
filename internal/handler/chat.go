package handler

import (
	"context"
	"io"
	"log/slog"

	"github.com/VaneZ444/chat-service/internal/usecase"
	pb "github.com/VaneZ444/golang-forum-protos/gen/go/chat" // путь к сгенерированному коду
)

type ChatServer struct {
	pb.UnimplementedChatServiceServer
	uc usecase.ChatUseCase
}

func NewChatServer(uc usecase.ChatUseCase) *ChatServer {
	return &ChatServer{uc: uc}
}

func (s *ChatServer) Connect(stream pb.ChatService_ConnectServer) error {
	msgChan, unsubscribe := s.uc.SubscribeToMessages()
	defer unsubscribe()

	// --- Отправка истории (например, 20 последних сообщений) ---
	history, err := s.uc.GetHistory(stream.Context(), 20)
	if err == nil {
		for _, m := range history {
			// можно оставить AuthorNickname пустым или добавить, если используешь кэш пользователей
			stream.Send(&pb.ServerMessage{
				Id:         m.ID,
				AuthorId:   m.AuthorID,
				AuthorName: m.AuthorNickname, // если есть поле в entity.Message
				Content:    m.Content,
				CreatedAt:  m.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	// --- Канал приёма сообщений от клиента ---
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF || err != nil {
				return
			}

			// Получаем user_id и nickname из metadata
			userID := GetUserIDFromCtx(stream.Context())
			if userID == 0 {
				userID = 1 // fallback для теста
			}
			nickname := GetUserNicknameFromCtx(stream.Context())

			_ = s.uc.SendMessage(stream.Context(), userID, nickname, msg.Content)

			// Можно сразу логировать
			slog.Info("new message received",
				"user_id", userID,
				"nickname", nickname,
				"content", msg.Content)
		}
	}()

	// --- Отдаём новые сообщения клиенту ---
	for {
		select {
		case m := <-msgChan:
			authorNickname := m.AuthorNickname
			if authorNickname == "" {
				authorNickname = "Unknown"
			}
			stream.Send(&pb.ServerMessage{
				Id:         m.ID,
				AuthorId:   m.AuthorID,
				AuthorName: authorNickname,
				Content:    m.Content,
				CreatedAt:  m.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (s *ChatServer) GetHistory(ctx context.Context, req *pb.HistoryRequest) (*pb.HistoryResponse, error) {
	messages, err := s.uc.GetHistory(ctx, int(req.Limit))
	if err != nil {
		return nil, err
	}

	var res pb.HistoryResponse
	for _, m := range messages {
		// Убедитесь, что AuthorNickname не пустой
		authorName := m.AuthorNickname
		if authorName == "" {
			authorName = "Аноним" // Запасное значение
		}

		res.Messages = append(res.Messages, &pb.ServerMessage{
			Id:         m.ID,
			AuthorId:   m.AuthorID,
			AuthorName: authorName, // Убедитесь, что это поле заполнено
			Content:    m.Content,
			CreatedAt:  m.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &res, nil
}
