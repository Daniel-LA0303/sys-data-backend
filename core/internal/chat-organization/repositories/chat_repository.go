package chat_repository

import (
	"context"
	chat_dto "core/project/core/internal/chat-organization/dtos"
	"database/sql"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error // Para listas
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row              // Para RETURNING
}

// get chats by user with last message and order
func (r *Repository) GetChatsByUser(
	ctx context.Context,
	userId string,
) ([]chat_dto.ChatsRoomsByUser, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
	SELECT 
    crt.chat_room_id,
    -- get display name group or user name
    CASE 
        WHEN crt.room_type = 'GROUP' THEN crt.name
        ELSE (
            SELECT u.username
            FROM room_members_tbl rm
            JOIN users_tbl u ON u.user_id = rm.user_id
            WHERE rm.chat_room_id = crt.chat_room_id
              AND rm.user_id <> $1
            LIMIT 1
        )
    END AS display_name,

    crt.name,
    crt.room_type,

    -- get last message
    m.messages_id,
    m.content,
    m.message_type,
    m.created_at AS last_message_date,

    -- get info sender
    sender.user_id AS sender_id,
    sender.username AS sender_username

	FROM chat_rooms_tbl crt

	-- filter by user participation
	LEFT JOIN room_members_tbl rm_self
		ON rm_self.chat_room_id = crt.chat_room_id
		AND rm_self.user_id = $1

	LEFT JOIN room_members_tbl rm_direct
		ON rm_direct.chat_room_id = crt.chat_room_id
		AND crt.room_type = 'DIRECT'
		AND rm_direct.user_id = $1

	-- get last message
	LEFT JOIN LATERAL (
		SELECT *
		FROM messages_tbl m
		WHERE m.room_id = crt.chat_room_id
		ORDER BY m.created_at DESC
		LIMIT 1
	) m ON TRUE

	LEFT JOIN users_tbl sender
		ON sender.user_id = m.user_id

	-- get only rooms with user that has joined
	WHERE (rm_self.user_id IS NOT NULL OR rm_direct.user_id IS NOT NULL)

	-- order 
	ORDER BY COALESCE(m.created_at, crt.created_at) DESC;
	`

	var rooms []chat_dto.ChatsRoomsByUser

	err := r.db.SelectContext(ctx, &rooms, query, userId)
	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *Repository) GetMessagesByRoomRepository(
	ctx context.Context,
	roomId string,
) ([]chat_dto.ChatMessage, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
	select 
		mt.messages_id,
		mt.room_id,
		mt.user_id,
		mt.content,
		mt.message_type,
		mt.created_at,
		mt.updated_at,
		mt.is_edited,
		ut.username
	from messages_tbl mt
	inner join users_tbl ut
		on ut.user_id = mt.user_id
	where mt.room_id = $1
	order by mt.created_at asc
	`

	var messages []chat_dto.ChatMessage

	err := r.db.SelectContext(ctx, &messages, query, roomId)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *Repository) CreateMessageRepository(
	ctx context.Context,
	o chat_dto.CreateMessageRequest,
) (*chat_dto.CreateMessageResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
	INSERT INTO messages_tbl (
		room_id,
		user_id,
		content,
		message_type,
		created_at,
		updated_at,
		is_edited
	)
	VALUES ($1, $2, $3, $4, NOW(), NULL, FALSE)
	RETURNING 
		messages_id,
		room_id,
		user_id,
		content,
		message_type,
		created_at,
		updated_at,
		is_edited;
	`

	var response chat_dto.CreateMessageResponse

	err := r.db.QueryRowContext( // ✅ usar r.db
		ctx,
		query,
		o.RoomId,
		o.UserId,
		o.Content,
		o.MessageType,
	).Scan(
		&response.MessagesId,
		&response.RoomId,
		&response.UserId,
		&response.Content,
		&response.MessageType,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.IsEdited,
	)

	if err != nil {
		log.Printf("INSERT MESSAGE ERROR: %v\n", err)
		return nil, err
	}

	return &response, nil
}

// create room
func (r *Repository) CreateChatRoom(
	ctx context.Context,
	tx *sqlx.Tx,
	room chat_dto.CreateChatRoomDB,
) (string, error) {

	query := `
	INSERT INTO chat_rooms_tbl (
		org_id,
		name,
		description,
		room_type,
		created_by
	)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING chat_room_id
	`

	var roomId string

	err := tx.GetContext(
		ctx,
		&roomId,
		query,
		room.OrgId,
		room.Name,
		room.Description,
		room.RoomType,
		room.UserIdCreator,
	)

	if err != nil {
		return "", err
	}

	return roomId, nil
}

// inserte member in room
func (r *Repository) InsertRoomMember(
	ctx context.Context,
	tx *sqlx.Tx,
	roomId string,
	userId string,
	role string,
) error {

	query := `
	INSERT INTO room_members_tbl (
		chat_room_id,
		user_id,
		role
	)
	VALUES ($1,$2,$3)
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		roomId,
		userId,
		role,
	)

	return err
}

// get chat info when has been created
func (r *Repository) GetChatRoomById(
	ctx context.Context,
	roomId string,
	userId string,
) (*chat_dto.ChatsRoomsByUser, error) {

	// this query is same as the one in GetChatsRoomsByUser but returns only one room
	query := `
    SELECT 
        crt.chat_room_id,
        CASE 
            WHEN crt.room_type = 'GROUP' THEN crt.name
            ELSE (
                SELECT u.username
                FROM room_members_tbl rm
                JOIN users_tbl u ON u.user_id = rm.user_id
                WHERE rm.chat_room_id = crt.chat_room_id
                  AND rm.user_id <> $2
                LIMIT 1
            )
        END AS display_name,
        crt.name,
        crt.room_type,
        m.messages_id,
        m.content,
        m.message_type,
        m.created_at AS last_message_date,
        sender.user_id AS sender_id,
        sender.username AS sender_username
    FROM chat_rooms_tbl crt
    LEFT JOIN LATERAL (
        SELECT * FROM messages_tbl m
        WHERE m.room_id = crt.chat_room_id
        ORDER BY m.created_at DESC LIMIT 1
    ) m ON TRUE
    LEFT JOIN users_tbl sender ON sender.user_id = m.user_id
    WHERE crt.chat_room_id = $1
    `
	var room chat_dto.ChatsRoomsByUser
	err := r.db.GetContext(ctx, &room, query, roomId, userId)
	if err != nil {
		return nil, err
	}

	return &room, nil
}
