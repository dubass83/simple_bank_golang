// Code generated by sqlc. DO NOT EDIT.

package db

import (
	"database/sql"
	"time"
)

type Account struct {
	ID        int64     `json:"id"`
	Owner     string    `json:"owner"`
	Balance   int64     `json:"balance"`
	Carrency  string    `json:"carrency"`
	CreatedAt time.Time `json:"createdAt"`
}

type Entry struct {
	ID        int64         `json:"id"`
	AccountID sql.NullInt64 `json:"accountID"`
	// can be negative or posetive
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"createdAt"`
}

type Transfer struct {
	ID            int64         `json:"id"`
	FromAccountID sql.NullInt64 `json:"fromAccountID"`
	ToAccountID   sql.NullInt64 `json:"toAccountID"`
	// should be posetive
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"createdAt"`
}

type User struct {
	Username          string    `json:"username"`
	HashedPassword    string    `json:"hashedPassword"`
	FullName          string    `json:"fullName"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"passwordChangedAt"`
	CreatedAt         time.Time `json:"createdAt"`
}
