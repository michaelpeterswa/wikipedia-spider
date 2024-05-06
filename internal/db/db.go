package db

import (
	"context"
	"errors"
	"fmt"

	_ "embed"

	"github.com/exaring/otelpgx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

type DBClient struct {
	Pool *pgxpool.Pool
}

//go:embed queries/insert_job.pgsql
var insertJobQuery string

//go:embed queries/insert_link.pgsql
var insertLinkQuery string

//go:embed queries/get_link_from_url.pgsql
var getLinkFromURLQuery string

func NewDBClient(ctx context.Context, connString string) (*DBClient, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DBClient{Pool: pool}, nil
}

func (c *DBClient) Close() {
	c.Pool.Close()
}

type Link struct {
	ID    *uuid.UUID
	JobID *uuid.UUID
	Link  string
	Title string
}

func (c *DBClient) GetLinkFromURL(ctx context.Context, url string) (*Link, error) {
	row := c.Pool.QueryRow(ctx, getLinkFromURLQuery, url)

	var link Link
	err := row.Scan(&link.ID, &link.JobID, &link.Link, &link.Title)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get link from url: %w", err)
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return &link, nil
}

func (c *DBClient) InsertLink(
	ctx context.Context,
	id *uuid.UUID,
	job_id *uuid.UUID,
	link string,
	title string,
) error {
	_, err := c.Pool.Exec(ctx, insertLinkQuery, id, job_id, link, title)
	if err != nil {
		return fmt.Errorf("insert link: %w", err)
	}

	return nil
}

func (c *DBClient) CreateJob(ctx context.Context) (*uuid.UUID, error) {
	jobID := uuid.New()
	_, err := c.Pool.Exec(ctx, insertJobQuery, jobID, "started")
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return &jobID, nil
}
