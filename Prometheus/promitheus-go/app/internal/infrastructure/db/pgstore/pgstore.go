package pgstore

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/entities/linkentity"
	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/usecase/app/repo"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var _ repo.LinkeStore = &Links{}

type DBPgLink struct {
	LinkID     uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
	OriginLink string
	ResultLink string
	LinkAt     time.Time
	Rank       int
}

type Links struct {
	db *sql.DB
}

func NewLinks(dsn string) (*Links, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS links (
		id uuid NOT NULL,
		created_at timestamptz NOT NULL,
		updated_at timestamptz NOT NULL,
		deleted_at timestamptz NULL,
		originLink varchar NOT NULL,
		resultLink varchar NOT NULL,
		link_at timestamptz NOT NULL,
		rank integer,
		CONSTRAINT links_pk PRIMARY KEY (id)
	)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	ls := &Links{
		db: db,
	}
	return ls, nil
}

func (ls *Links) Close() {
	ls.db.Close()
}

func (ls *Links) Create(ctx context.Context, l linkentity.Link) (*uuid.UUID, error) {
	dbu := &DBPgLink{
		LinkID:     l.LinkID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		OriginLink: l.OriginLink,
		ResultLink: l.ResultLink,
		LinkAt:     time.Now(),
	}

	_, err := ls.db.ExecContext(ctx, `INSERT INTO links 
	(id, created_at, updated_at, deleted_at, originLink, resultLink, link_at, rank )
	values ($1, $2, $3, $4, $5, $6, $7, $8)`,
		dbu.LinkID,
		dbu.CreatedAt,
		dbu.UpdatedAt,
		nil,
		dbu.OriginLink,
		dbu.ResultLink,
		dbu.LinkAt,
		dbu.Rank,
	)
	if err != nil {
		return nil, err
	}

	return &l.LinkID, nil
}

func (ls *Links) Delete(ctx context.Context, uid uuid.UUID) error {
	_, err := ls.db.ExecContext(ctx, `UPDATE links SET deleted_at = $2 WHERE id = $1`,
		uid, time.Now(),
	)
	return err
}

func (ls *Links) ReadLinkRank(ctx context.Context, uid uuid.UUID) (*linkentity.Link, error) {
	dbu := &DBPgLink{}
	rows, err := ls.db.QueryContext(ctx,
		`SELECT id, created_at, updated_at, deleted_at, originLink, resultLink, link_at, rank
	FROM users WHERE id = $1`, uid)
	if err != nil {
		return nil, rows.Err()
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(
			&dbu.LinkID,
			&dbu.CreatedAt,
			&dbu.UpdatedAt,
			&dbu.DeletedAt,
			&dbu.OriginLink,
			&dbu.ResultLink,
			&dbu.LinkAt,
			&dbu.Rank,
		); err != nil {
			return nil, err
		}
	}

	return &linkentity.Link{
		LinkID:     dbu.LinkID,
		OriginLink: dbu.OriginLink,
		ResultLink: dbu.ResultLink,
		LinkAt:     dbu.LinkAt,
		Rank:       dbu.Rank,
	}, nil
}

func (ls *Links) SearchLinks(ctx context.Context, s string) (chan linkentity.Link, error) {
	const buf = 100
	chout := make(chan linkentity.Link, buf)

	go func() {
		defer close(chout)
		dbu := &DBPgLink{}

		rows, err := ls.db.QueryContext( //nolint
			ctx, `
		SELECT id, created_at, updated_at, originLink, resultLink, link_at, rank
		FROM users WHERE name LIKE $1 and deleted_at is null`, "%"+s+"%")
		if err != sql.ErrNoRows { //nolint
			return
		}
		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(
				&dbu.LinkID,
				&dbu.CreatedAt,
				&dbu.UpdatedAt,
				&dbu.DeletedAt,
				&dbu.OriginLink,
				&dbu.ResultLink,
				&dbu.LinkAt,
				&dbu.Rank,
			); err != nil {
				log.Println(err)
				return
			}

			chout <- linkentity.Link{
				LinkID:     dbu.LinkID,
				OriginLink: dbu.OriginLink,
				ResultLink: dbu.ResultLink,
				LinkAt:     dbu.LinkAt,
				Rank:       dbu.Rank,
			}
		}
	}()

	return chout, nil
}

//функция для удовлетворения interface
func (ls *Links) GetLongURL(ctx context.Context, sh string) (string, error) {
	return sh, nil
}

func (ls *Links) RankCounter(ctx context.Context, uid uuid.UUID, rank int) error {
	_, err := ls.db.ExecContext(ctx, `UPDATE links SET rank = $2 WHERE id = $1`,
		uid, rank,
	)
	return err
}
