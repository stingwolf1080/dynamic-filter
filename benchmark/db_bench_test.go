package benchmark

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stingwolf1080/dynamic-filter/benchmark/sqlc_gen"
	"github.com/stingwolf1080/dynamic-filter/benchmark/sqlx_gen"
)

func setupMockDB(b *testing.B) (sqlmock.Sqlmock, *sqlc_gen.Queries, *sqlx_gen.Queries) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		b.Fatalf("failed to open sqlmock: %s", err)
	}

	sqlxDB := sqlx.NewDb(db, "postgres")
	sqlcQueries := sqlc_gen.New(db)
	sqlxQueries := sqlx_gen.New(sqlxDB)

	return mock, sqlcQueries, sqlxQueries
}

func BenchmarkSqlc_GetUser(b *testing.B) {
	mock, sqlcQueries, _ := setupMockDB(b)

	now := time.Now()
	
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		mock.ExpectQuery(`SELECT id, username, email, created_at FROM users WHERE id = \$1 LIMIT 1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "created_at"}).AddRow(1, "testuser", "test@example.com", now))
		b.StartTimer()

		_, err := sqlcQueries.GetUser(ctx, 1)
		if err != nil {
			b.Fatalf("sqlc GetUser failed: %v", err)
		}
	}
}

func BenchmarkSqlx_GetUser(b *testing.B) {
	mock, _, sqlxQueries := setupMockDB(b)

	now := time.Now()
	
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		mock.ExpectQuery(`SELECT id, username, email, created_at FROM users WHERE id = \$1 LIMIT 1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "created_at"}).AddRow(1, "testuser", "test@example.com", now))
		b.StartTimer()

		_, err := sqlxQueries.GetUser(ctx, 1)
		if err != nil {
			b.Fatalf("sqlx GetUser failed: %v", err)
		}
	}
}

func BenchmarkSqlc_ListUsers(b *testing.B) {
	mock, sqlcQueries, _ := setupMockDB(b)
	now := time.Now()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rows := sqlmock.NewRows([]string{"id", "username", "email", "created_at"})
		for j := 0; j < 100; j++ {
			rows.AddRow(j, "user", "test@test.com", now)
		}
		mock.ExpectQuery(`SELECT id, username, email, created_at FROM users ORDER BY id LIMIT \$1 OFFSET \$2`).
			WithArgs(100, 0).
			WillReturnRows(rows)
		b.StartTimer()

		_, err := sqlcQueries.ListUsers(ctx, sqlc_gen.ListUsersParams{Limit: 100, Offset: 0})
		if err != nil {
			b.Fatalf("sqlc ListUsers failed: %v", err)
		}
	}
}

func BenchmarkSqlx_ListUsers(b *testing.B) {
	mock, _, sqlxQueries := setupMockDB(b)
	now := time.Now()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rows := sqlmock.NewRows([]string{"id", "username", "email", "created_at"})
		for j := 0; j < 100; j++ {
			rows.AddRow(j, "user", "test@test.com", now)
		}
		mock.ExpectQuery(`SELECT id, username, email, created_at FROM users ORDER BY id LIMIT \$1 OFFSET \$2`).
			WithArgs(100, 0).
			WillReturnRows(rows)
		b.StartTimer()

		_, err := sqlxQueries.ListUsers(ctx, 100, 0)
		if err != nil {
			b.Fatalf("sqlx ListUsers failed: %v", err)
		}
	}
}
