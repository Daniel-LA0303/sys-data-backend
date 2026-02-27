package chat_dto

import "time"

// we use pointers to avoid nulls
type ChatsRoomsByUser struct {
	ChatRoomId      string     `json:"chatRoomId" db:"chat_room_id"`
	DisplayName     string     `json:"displayName" db:"display_name"`
	RoomType        string     `json:"roomType" db:"room_type"`
	MessagesId      *string    `json:"messagesId" db:"messages_id"`
	Content         *string    `json:"content" db:"content"`
	MessageType     *string    `json:"messageType" db:"message_type"`
	LastMessageDate *time.Time `json:"lastMessageDate" db:"last_message_date"`
	SenderId        *string    `json:"senderId" db:"sender_id"`
	SenderUsername  *string    `json:"senderUsername" db:"sender_username"`
}

type ChatMessage struct {
	MessagesId  string     `json:"messagesId" db:"messages_id"`
	RoomId      string     `json:"roomId" db:"room_id"`
	UserId      string     `json:"userId" db:"user_id"`
	Content     string     `json:"content" db:"content"`
	MessageType string     `json:"messageType" db:"message_type"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   *time.Time `json:"updatedAt" db:"updated_at"`
	IsEdited    bool       `json:"isEdited" db:"is_edited"`
	Username    string     `json:"username" db:"username"`
}

type CreateMessageRequest struct {
	RoomId      string `json:"roomId" db:"room_id"`
	UserId      string `json:"userId" db:"user_id"`
	Content     string `json:"content" db:"content"`
	MessageType string `json:"messageType" db:"message_type"`
}

type CreateMessageResponse struct {
	MessagesId  string  `json:"messagesId" db:"messages_id"`
	RoomId      string  `json:"roomId" db:"room_id"`
	UserId      string  `json:"userId" db:"user_id"`
	Content     string  `json:"content" db:"content"`
	MessageType string  `json:"messageType" db:"message_type"`
	CreatedAt   string  `json:"createdAt" db:"created_at"`
	UpdatedAt   *string `json:"updatedAt" db:"updated_at"`
	IsEdited    bool    `json:"isEdited" db:"is_edited"`
}
