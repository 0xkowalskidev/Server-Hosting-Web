package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameserverFormData struct {
	Name   string
	GameID string
}

type Gameserver struct {
	gorm.Model
	ID   uuid.UUID `gorm:"type:uuid;primary_key;"`
	Name string

	// Belongs to Game
	GameID uuid.UUID
	Game   Game

	// Resource Limits
	StorageLimit int
	MemoryLimit  int
}

// Set unique ID
func (server *Gameserver) BeforeCreate(tx *gorm.DB) (err error) {
	if server.ID == uuid.Nil {
		server.ID = uuid.New()
	}
	return
}
