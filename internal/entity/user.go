package entity

import "time"

const UserMaxLoginLen = 64

type User struct {
	ID      uint64
	Login   string
	Hash    string
	Created time.Time
}
