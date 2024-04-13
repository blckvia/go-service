package models

import (
	"errors"
	"time"
)

type Goods struct {
	ID          int       `json:"id" db:"id"`
	ProjectID   int       `json:"project_id" db:"project_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Priority    int       `json:"priority" db:"priority"`
	Removed     bool      `json:"removed" db:"removed"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type UpdateGoods struct {
	Name        *string `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
}

type GetAllGoodsResponse struct {
	Meta  Meta    `json:"meta"`
	Goods []Goods `json:"goods"`
}

type Meta struct {
	Total   int `json:"total"`
	Removed int `json:"removed"`
	Limit   int `json:"limit"`
	Offset  int `json:"offset"`
}

func (i UpdateGoods) Validate() error {
	if i.Description == nil {
		return errors.New("empty description")
	}
	return nil
}
