package chat_services

import (
	"context"
	chat_dto "core/project/core/internal/chat-organization/dtos"
	chat_repository "core/project/core/internal/chat-organization/repositories"
	"log"
)

type Service struct {
	repo *chat_repository.Repository
}

func NewService(repo *chat_repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetChatsByUser(ctx context.Context, userId string, orgId string) ([]chat_dto.ChatsRoomsByUser, error) {
	// Validaciones básicas
	log.Printf("GetChatsByUser: %v %v\n", userId)
	return s.repo.GetChatsByUser(ctx, userId)
}

func (s *Service) GetMessagesByRoom(ctx context.Context, roomId string) ([]chat_dto.ChatMessage, error) {
	// Validaciones básicas

	return s.repo.GetMessagesByRoomRepository(ctx, roomId)
}

func (s *Service) CreateMessage(
	ctx context.Context,
	o chat_dto.CreateMessageRequest,
) (*chat_dto.CreateMessageResponse, error) {

	// Aquí puedes meter validaciones después
	return s.repo.CreateMessageRepository(ctx, o)
}
