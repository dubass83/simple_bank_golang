// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: user.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (
  username, hashed_password, full_name, email
) VALUES (
  $1, $2, $3, $4
)
RETURNING username, hashed_password, full_name, email, password_changed_at, created_at, is_email_verified
`

type CreateUserParams struct {
	Username       string `json:"username"`
	HashedPassword string `json:"hashedPassword"`
	FullName       string `json:"fullName"`
	Email          string `json:"email"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.Username,
		arg.HashedPassword,
		arg.FullName,
		arg.Email,
	)
	var i User
	err := row.Scan(
		&i.Username,
		&i.HashedPassword,
		&i.FullName,
		&i.Email,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.IsEmailVerified,
	)
	return i, err
}

const getUser = `-- name: GetUser :one
SELECT username, hashed_password, full_name, email, password_changed_at, created_at, is_email_verified FROM users
WHERE username = $1 LIMIT 1
`

func (q *Queries) GetUser(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRow(ctx, getUser, username)
	var i User
	err := row.Scan(
		&i.Username,
		&i.HashedPassword,
		&i.FullName,
		&i.Email,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.IsEmailVerified,
	)
	return i, err
}

const updateUser = `-- name: UpdateUser :one
UPDATE users 
SET 
  hashed_password = COALESCE($1, hashed_password),
  full_name = COALESCE($2, full_name),
  email = COALESCE($3, email),
  is_email_verified = COALESCE($4, is_email_verified),
  password_changed_at = COALESCE($5, password_changed_at)
WHERE 
  username = $6
RETURNING username, hashed_password, full_name, email, password_changed_at, created_at, is_email_verified
`

type UpdateUserParams struct {
	HashedPassword    pgtype.Text        `json:"hashedPassword"`
	FullName          pgtype.Text        `json:"fullName"`
	Email             pgtype.Text        `json:"email"`
	IsEmailVerified   pgtype.Bool        `json:"isEmailVerified"`
	PasswordChangedAt pgtype.Timestamptz `json:"passwordChangedAt"`
	Username          string             `json:"username"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUser,
		arg.HashedPassword,
		arg.FullName,
		arg.Email,
		arg.IsEmailVerified,
		arg.PasswordChangedAt,
		arg.Username,
	)
	var i User
	err := row.Scan(
		&i.Username,
		&i.HashedPassword,
		&i.FullName,
		&i.Email,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.IsEmailVerified,
	)
	return i, err
}
