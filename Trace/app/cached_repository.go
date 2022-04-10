package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type cachedRepository struct {
	repository Repository
	logger     *zap.Logger
	cache      *cache.Cache
}

func (r *cachedRepository) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	key := fmt.Sprintf("user:%s", id)
	var user User
	err := r.cache.Get(ctx, key, &user)
	switch err {
	case nil:
		return &user, nil
	case cache.ErrCacheMiss:
		dbUser, dbErr := r.repository.GetUser(ctx, id)
		if dbErr != nil {
			return nil, dbErr
		}
		err = r.cache.Set(&cache.Item{
			Ctx:   ctx,
			Key:   key,
			Value: dbUser,
			TTL:   time.Hour,
		})

		if err != nil {
			return nil, err
		}
		return dbUser, nil
	}
	return nil, err
}

func (r *cachedRepository) GetUsersByName(ctx context.Context, name string) ([]*User, error) {
	r.logger.Info("in get users by name")
	key := fmt.Sprintf("users-name:%s", name)
	var users []*User
	err := r.cache.Get(ctx, key, &users)
	switch err {
	case nil:
		return users, nil
	case cache.ErrCacheMiss:
		r.logger.Info("cache miss!")
		dbUsers, dbErr := r.repository.GetUsersByName(ctx, name)
		if dbErr != nil {
			return nil, dbErr
		}
		err = r.cache.Set(&cache.Item{
			Ctx:   ctx,
			Key:   key,
			Value: dbUsers,
			TTL:   time.Second * 5,
		})
		if err != nil {
			return nil, err
		}
		return dbUsers, nil
	}
	return nil, err
}

func (r *cachedRepository) GetUsers(ctx context.Context) ([]User, error) {
	return r.repository.GetUsers(ctx)
}

func (r *cachedRepository) GetUserArticles(ctx context.Context, userID uuid.UUID) ([]Article, error) {
	key := fmt.Sprintf("user_articles: %s", userID)
	var articles []Article
	err := r.cache.Get(ctx, key, &articles)
	switch err {
	case nil:
		return articles, nil
	case cache.ErrCacheMiss:

		dbArticles, dbErr := r.repository.GetUserArticles(ctx, userID)
		if dbErr != nil {
			return nil, dbErr
		}
		err = r.cache.Set(&cache.Item{
			Ctx:   ctx,
			Key:   key,
			Value: dbArticles,
			TTL:   time.Second * 5,
		})
		if err != nil {
			return nil, err
		}
		return dbArticles, nil
	}
	return nil, err
}

func NewCachedRepository(repository Repository, logger *zap.Logger) *cachedRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	rCache := cache.New(&cache.Options{
		Redis:      rdb,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})
	return &cachedRepository{
		repository: repository,
		cache:      rCache,
		logger:     logger,
	}
}
