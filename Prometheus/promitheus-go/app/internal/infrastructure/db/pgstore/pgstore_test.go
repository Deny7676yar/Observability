package pgstore

import (
	"database/sql"
	"os"
	"testing"

	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/test"
	gc "gopkg.in/check.v1"
)

var _ = gc.Suite(new(PgTestSuite))

type PgTestSuite struct {
	test.SuitBase
	db *sql.DB
}

func Test(t *testing.T) {
	gc.TestingT(t)
}

func (s *PgTestSuite) SetUpSuite(c *gc.C) {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		c.Skip("Missing PGDB_DSN envvar; skipping postgresdb-backed test suite")
	}

	l, err := NewLinks(dsn)
	c.Assert(err, gc.IsNil)
	s.db = l.db
}

func (s *PgTestSuite) SetUpTest(c *gc.C) {
	s.flushDB(c)
}

func (s *PgTestSuite) TearDownSuite(c *gc.C) {
	if s.db != nil {
		s.flushDB(c)
		c.Assert(s.db.Close(), gc.IsNil)
	}
}

func (s *PgTestSuite) flushDB(c *gc.C) {
	_, err := s.db.Exec("DELETE FROM links")
	c.Assert(err, gc.IsNil)
}
