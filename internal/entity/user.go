package entity

import "time"

const UserMaxLoginLen = 64

type User struct {
	ID      uint64    `db:"id"`
	Login   string    `db:"login"`
	Hash    string    `db:"hash"`
	Created time.Time `db:"created_at"`
}
