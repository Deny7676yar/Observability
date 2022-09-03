package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/entities/linkentity"
	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/usecase/app/repo"
	"github.com/google/uuid"
)

type Handlers struct {
	ls *repo.Links
}

func NewHandlers(ls *repo.Links) *Handlers {
	r := &Handlers{
		ls: ls,
	}
	return r
}

type Link struct {
	LinkID     uuid.UUID `json:"linkId"`
	OriginLink string    `json:"originLink"`
	ResultLink string    `json:"resultLink"`
	LinkAt     time.Time `json:"linkAt"`
	Rank       int       `json:"rank"`
}

func (rt *Handlers) CreateLink(ctx context.Context, l Link) (Link, error) {
	// DTO
	bu := linkentity.Link{
		OriginLink: l.OriginLink,
		ResultLink: l.ResultLink,
	}

	nbu, err := rt.ls.Create(ctx, bu)
	if err != nil {
		return Link{}, fmt.Errorf("error when creating: %w", err)
	}

	return Link{
		LinkID:     nbu.LinkID,
		OriginLink: nbu.OriginLink,
		ResultLink: nbu.ResultLink,
		LinkAt:     nbu.LinkAt,
		Rank:       nbu.Rank,
	}, nil
}

var ErrUserNotFound = errors.New("user not found")

// /read?uid=...
func (rt *Handlers) ReadLinkRank(ctx context.Context, uid uuid.UUID) (Link, error) {
	if (uid == uuid.UUID{}) {
		return Link{}, fmt.Errorf("Read, bad request: uid is empty") //nolint
	}

	nbu, err := rt.ls.ReadLinkRank(ctx, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Link{}, ErrUserNotFound
		}
		return Link{}, fmt.Errorf("error when reading: %w", err)
	}

	return Link{
		LinkID:     nbu.LinkID,
		OriginLink: nbu.OriginLink,
		ResultLink: nbu.ResultLink,
		LinkAt:     nbu.LinkAt,
		Rank:       nbu.Rank,
	}, nil
}

func (rt *Handlers) DeleteLink(ctx context.Context, uid uuid.UUID) (Link, error) {
	if (uid == uuid.UUID{}) {
		return Link{}, fmt.Errorf("Delete, bad request: uid is empty") //nolint
	}

	nbu, err := rt.ls.Delete(ctx, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Link{}, ErrUserNotFound
		}
		return Link{}, fmt.Errorf("error when delete: %w", err)
	}

	return Link{
		LinkID:     nbu.LinkID,
		OriginLink: nbu.OriginLink,
		ResultLink: nbu.ResultLink,
		LinkAt:     nbu.LinkAt,
	}, nil
}

// /search?q=...
func (rt *Handlers) SearchLink(ctx context.Context, q string, f func(Link) error) error {
	ch, err := rt.ls.SearchLinks(ctx, q)
	if err != nil {
		return fmt.Errorf("error when reading: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case l, ok := <-ch:
			if !ok {
				return nil
			}
			if err := f(Link{
				LinkID:     l.LinkID,
				OriginLink: l.OriginLink,
				ResultLink: l.ResultLink,
				LinkAt:     l.LinkAt,
				Rank:       l.Rank,
			}); err != nil {
				return err
			}
		}
	}
}

// /visitor?shortURL=...
func (rt *Handlers) GetLongURL(ctx context.Context, sh string) (string, error) {
	longURL, err := rt.ls.GetLongURL(ctx, sh)
	if err != nil {
		return "", err
	}
	for {
		select {
		case <-ctx.Done():
			return "", err
		default:
			return longURL, nil
		}
	}
}
