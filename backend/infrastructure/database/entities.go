package database

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoomStatusMatching = "matching"
	RoomStatusPlaying  = "playing"
	RoomStatusFinished = "finished"
)

type RoomEntity struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Status    string    `gorm:"type:text;not null"`
	StartAt   *time.Time
	ExpiresAt *time.Time
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`

	Players []PlayerEntity `gorm:"foreignKey:RoomID;references:ID"`
}

func (RoomEntity) TableName() string {
	return "rooms"
}

type PlayerEntity struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	RoomID     uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:uq_room_order"`
	Name       string    `gorm:"type:varchar(255);not null"`
	Location   string    `gorm:"type:geography(Point,4326);not null"`
	OrderIndex int       `gorm:"not null;check:order_index >= 0;uniqueIndex:uq_room_order"`
	CreatedAt  time.Time `gorm:"not null;default:now()"`
	UpdatedAt  time.Time `gorm:"not null;default:now()"`
}

func (PlayerEntity) TableName() string {
	return "players"
}

type MessageEntity struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	RoomID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	SenderID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	ReceiverID *uuid.UUID `gorm:"type:uuid;index"`
	Content    string     `gorm:"type:text;not null"`
	CreatedAt  time.Time  `gorm:"not null;default:now()"`
}

func (MessageEntity) TableName() string {
	return "messages"
}

type ConnectionEntity struct {
	RoomID     uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	SenderID   uuid.UUID `gorm:"type:uuid;not null;primaryKey;index"`
	ReceiverID uuid.UUID `gorm:"type:uuid;not null;primaryKey;index"`
	NeedsSwap  bool      `gorm:"not null;default:false"`
	CreatedAt  time.Time `gorm:"not null;default:now()"`
	UpdatedAt  time.Time `gorm:"not null;default:now()"`
}

func (ConnectionEntity) TableName() string {
	return "connections"
}
