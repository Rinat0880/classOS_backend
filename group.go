package classosbackend

import (
	"errors"
	"time"
)

type Group struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name" binding:"required"`
}

type WhitelistEntry struct {
	ID        int64     `json:"id"`
	GroupID   int64     `json:"group_id"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

type Settings struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateGroupInput struct {
	Name *string `json:"name"`
}

func (i UpdateGroupInput) Validate() error {
    if i.Name == nil {
        return errors.New("update structure has no values")
    }

    return nil
}
