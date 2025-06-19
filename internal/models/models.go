package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username       string    `gorm:"uniqueIndex;not null" json:"username"`
	Email          string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash   string    `gorm:"not null" json:"-"`
	Bio            string    `json:"bio"`
	ProfilePicture string    `json:"profile_picture"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	gorm.Model
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SenderID   uuid.UUID `gorm:"type:uuid;not null;index" json:"sender_id"`
	ReceiverID uuid.UUID `gorm:"type:uuid;not null;index" json:"receiver_id"`
	Body       string    `gorm:"not null" json:"body"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RefreshToken struct {
	gorm.Model
	ID        uint
	Token     uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"token"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	IssuedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"issued_at"`
	User      User      `gorm:"foreignKey:UserID"` // GORM association
}
