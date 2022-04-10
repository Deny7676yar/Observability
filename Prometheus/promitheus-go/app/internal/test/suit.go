package test

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/entities/linkentity"
	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/usecase/app/repo"
	"github.com/google/uuid"
	gc "gopkg.in/check.v1"
)

type SuitBase struct {
	l repo.LinkeStore
}

func (s *SuitBase) Setlink(link repo.LinkeStore) {
	s.l = link
}

func (s *SuitBase) TestCreateLink(c *gc.C) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	origin := &linkentity.Link{
		OriginLink: "https://example.com",
		LinkAt:     time.Now().Add(-10 * time.Hour),
	}
	_, err := s.l.Create(ctx, *origin)
	c.Assert(err, gc.IsNil)
	c.Assert(origin.LinkID, gc.Not(gc.Equals), uuid.Nil, gc.Commentf(""))

	accessedAt := time.Now().Truncate(time.Second).UTC()
	existing := &linkentity.Link{
		LinkID:     origin.LinkID,
		OriginLink: "https://example.com",
		LinkAt:     accessedAt,
	}
	_, err = s.l.Create(ctx, *existing)
	c.Assert(err, gc.IsNil)
	c.Assert(existing.LinkID, gc.Equals, origin.LinkID, gc.Commentf("link ID changed while upserting"))
	cancel()
}
