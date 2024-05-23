package models

import (
	"github.com/uptrace/bun"
)

//go:generate generate/cars.py

type Car struct {
	bun.BaseModel

	Ordinal int    `json:"id" bun:",unique,pk"`
	Year    int    `json:"year"`
	Make    string `json:"make"`
	Model   string `json:"model"`
}

type CarClass struct {
	bun.BaseModel

	Id      int    `json:"id" bun:",unique,pk"`
	Name    string `json:"name"`
	PIStart int    `bun:"pi_start"`
	PIEnd   int    `bun:"pi_end"`
	Color   string `json:"color"`
}

var CarClasses = []CarClass{
	{
		Id:      0,
		Name:    "E",
		PIStart: 100,
		PIEnd:   300,
		Color:   "E91E63",
	},
	{
		Id:      1,
		Name:    "D",
		PIStart: 301,
		PIEnd:   400,
		Color:   "607D8B",
	},
	{
		Id:      2,
		Name:    "C",
		PIStart: 401,
		PIEnd:   500,
		Color:   "FF9800",
	},
	{
		Id:      3,
		Name:    "B",
		PIStart: 501,
		PIEnd:   600,
		Color:   "FF5722",
	},
	{
		Id:      4,
		Name:    "A",
		PIStart: 601,
		PIEnd:   700,
		Color:   "F44336",
	},
	{
		Id:      5,
		Name:    "S",
		PIStart: 701,
		PIEnd:   800,
		Color:   "9C27B0",
	},
	{
		Id:      6,
		Name:    "R",
		PIStart: 801,
		PIEnd:   900,
		Color:   "2196F3",
	},
	{
		Id:      7,
		Name:    "P",
		PIStart: 901,
		PIEnd:   998,
		Color:   "4CAF50",
	},
	{
		Id:      8,
		Name:    "X",
		PIStart: 999,
		PIEnd:   1000,
		Color:   "9E9E9E",
	},
}
