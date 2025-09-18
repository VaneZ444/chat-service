package handler

import (
	"context"
	"io"

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

	// Отправка истории (например, 50 последних сообщений)
	history, err := s.uc.GetHistory(stream.Context(), 50)
	if err == nil {
		for _, m := range history {
			stream.Send(&pb.ServerMessage{
				Id:        m.ID,
				AuthorId:  m.AuthorID,
				Content:   m.Content,
				CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	// Канал приёма сообщений от клиента
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF || err != nil {
				return
			}
			// user_id придёт из метаданных (gateway должен прокидывать)
			// пока можно хардкодить для теста
			_ = s.uc.SendMessage(stream.Context(), 1, msg.Content)
		}
	}()

	// Отдаём новые сообщения клиенту
	for {
		select {
		case m := <-msgChan:
			stream.Send(&pb.ServerMessage{
				Id:        m.ID,
				AuthorId:  m.AuthorID,
				Content:   m.Content,
				CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
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
		res.Messages = append(res.Messages, &pb.ServerMessage{
			Id:        m.ID,
			AuthorId:  m.AuthorID,
			Content:   m.Content,
			CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &res, nil
}
