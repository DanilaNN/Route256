//go:build integration

package integration_test

import (
	"context"
	"log"
	"route256/checkout/internal/domain"
	postgres "route256/checkout/internal/repository"
	"route256/checkout/internal/repository/postgress/tx"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/suite"
)

const pgTestDBA = "postgres://user:password@localhost:5435/test?sslmode=disable"

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite
	repo domain.PGRepository
	pool *pgxpool.Pool // to make simple SQL queries
}

func (s *Suite) SetupSuite() {
	log.Println("SetupSuite()")

	pool, err := pgxpool.Connect(context.Background(), pgTestDBA)
	s.Require().NoError(err)

	provider := tx.New(pool)
	s.repo = postgres.New(provider)
	s.pool = pool
}

func (s *Suite) TearDownSuite() {
	log.Println("TearDownSuite()")

	_, err := s.pool.Exec(context.Background(), `DELETE FROM carts`)
	s.Require().NoError(err)
	s.pool.Close()
}

func (s *Suite) SetupTest() {
	log.Println("SetupTest()")
	_, err := s.pool.Exec(context.Background(), `DELETE FROM carts`)
	s.Require().NoError(err)
}
