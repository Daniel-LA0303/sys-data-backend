package chat_services

import (
	"context"
	"core/project/core/common/errors"
	"core/project/core/common/validator"
	chat_dto "core/project/core/internal/chat-organization/dtos"
	chat_repository "core/project/core/internal/chat-organization/repositories"
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	repo *chat_repository.Repository
	tx   TxManager
	db   *sqlx.DB
}

type TxManager interface {
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

// Pasa el db (*sqlx.DB) para poder manejar transacciones
func NewService(repo *chat_repository.Repository, db *sqlx.DB) *Service {
	return &Service{
		repo: repo,
		tx:   db, // sqlx.DB cumple con la interfaz TxManager
		db:   db,
	}
}

func (s *Service) GetChatsByUser(ctx context.Context, userId string, orgId string) ([]chat_dto.ChatsRoomsByUser, error) {
	// Validaciones básicas
	log.Printf("GetChatsByUser: %v\n", userId)
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

func (s *Service) CreateRoom(
	ctx context.Context,
	input chat_dto.CreateChatRoomRequest,
) (chat_dto.CreateChatRoomResponse, error) {

	validationErrors := validator.ValidateStruct(input.ChatRoomInfo)
	if validationErrors != nil {
		return chat_dto.CreateChatRoomResponse{}, errors.NewValidationFiledsError("Invalid input", validationErrors)
	}

	// 1. init transaction
	tx, err := s.tx.BeginTxx(ctx, nil)
	if err != nil {
		return chat_dto.CreateChatRoomResponse{}, errors.NewDatabaseError(err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// 2. create room
	roomId, err := s.repo.CreateChatRoom(ctx, tx, input.ChatRoomInfo)
	if err != nil {
		return chat_dto.CreateChatRoomResponse{}, err
	}

	// 3. insert creator as owner
	err = s.repo.InsertRoomMember(
		ctx,
		tx,
		roomId,
		input.ChatRoomInfo.UserIdCreator,
		"ADMIN",
	)
	if err != nil {
		return chat_dto.CreateChatRoomResponse{}, err
	}

	// 4. insert members
	for _, userId := range input.ChatRoomMembersInfo.UserId {

		if userId == input.ChatRoomInfo.UserIdCreator {
			continue
		}

		err = s.repo.InsertRoomMember(
			ctx,
			tx,
			roomId,
			userId,
			"member",
		)

		if err != nil {
			return chat_dto.CreateChatRoomResponse{}, err
		}
	}
	// 5. commit
	if err = tx.Commit(); err != nil {
		return chat_dto.CreateChatRoomResponse{}, errors.NewDatabaseError(err)
	}

	// 6. get info room has been created
	roomData, err := s.repo.GetChatRoomById(ctx, roomId, input.ChatRoomInfo.UserIdCreator)
	if err != nil {
		return chat_dto.CreateChatRoomResponse{}, errors.NewDatabaseError(err)
	}

	res := chat_dto.CreateChatRoomResponse{
		Room: *roomData,
	}

	return res, nil
}
