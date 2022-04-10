package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	UsersSelect        = `SELECT id, name FROM users`
	UserByIDSelect     = `SELECT id, name FROM users WHERE id = $1`
	UserArticlesSelect = `SELECT id, title, text, user_id FROM articles WHERE user_id = $1`
)

var (
	ErrNotFound      = errors.New("not found")
	ErrMultipleFound = errors.New("multiple found")
)

type repository struct {
	pool *pgxpool.Pool
}

func (r *repository) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	rows, _ := r.pool.Query(ctx, UserByIDSelect, id)
	var (
		user  User
		found bool
	)
	for rows.Next() {
		if found {
			return nil, fmt.Errorf("%w: user id %s", ErrMultipleFound, id)
		}
		if err := rows.Scan(&user.ID, &user.Name); err != nil {
			return nil, err
		}
		found = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("%w: user id %s", ErrNotFound, id)
	}
	return &user, nil
}

func (r *repository) GetUsersByName(ctx context.Context, name string) ([]*User, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM users WHERE name = $1`, name)
	if err != nil {
		return nil, fmt.Errorf("DB query failed: %w", err)
	}
	defer rows.Close()
	users := make([]*User, 0)
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, fmt.Errorf("failed to parse the received result: %w", err)
		}
		users = append(users, &u)
	}
	return users, nil
}

func (r *repository) GetUsers(ctx context.Context) ([]User, error) {
	rows, _ := r.pool.Query(ctx, UsersSelect)
	ret := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name); err != nil {
			return nil, err
		}
		ret = append(ret, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (r *repository) GetUserArticles(ctx context.Context, userID uuid.UUID) ([]Article, error) {
	rows, _ := r.pool.Query(ctx, UserArticlesSelect, userID)
	ret := make([]Article, 0)
	for rows.Next() {
		var article Article
		if err := rows.Scan(&article.ID, &article.Title, &article.Text,
			&article.UserID); err != nil {
			return nil, err
		}
		ret = append(ret, article)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{pool: pool}
}
