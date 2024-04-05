package models

import (
	"errors"
	"time"
)

type Project struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type UpdateProject struct {
	Name *string `json:"name" db:"name"`
}

func (i UpdateProject) Validate() error {
	if i.Name == nil {
		return errors.New("empty name")
	}
	return nil
}
