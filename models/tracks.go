package models

import (
	"fmt"

	"github.com/uptrace/bun"
)

//go:generate generate/tracks.py

type Track struct {
	bun.BaseModel

	Ordinal  int     `json:"id" bun:",unique,pk"`
	Name     string  `json:"name"`
	Layout   string  `json:"layout"`
	Location string  `json:"location"`
	Length   float64 `json:"length"` // in kilometers
}

func (t Track) FullName() string {
	return fmt.Sprintf("%v - %v", t.Name, t.Layout)
}
