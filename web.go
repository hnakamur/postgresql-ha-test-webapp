package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type MyApp struct {
	pool *pgxpool.Pool
}

func newMyApp(ctx context.Context, databaseURL string) (*MyApp, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	return &MyApp{
		pool: pool,
	}, nil
}

func (a *MyApp) closeDB() {
	a.pool.Close()
}

func (a *MyApp) addHandlers(e *echo.Echo) {
	e.POST("/ping", a.handlePostPing)
	e.GET("/ping/latest", a.handleGetPingLatest)
}

type ping struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

func (a *MyApp) handlePostPing(c echo.Context) error {
	ctx := c.Request().Context()
	var p ping
	err := pgx.BeginFunc(ctx, a.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, "INSERT INTO pings (id, created_at) VALUES (DEFAULT, $1) RETURNING id, created_at", time.Now()).Scan(&p.ID, &p.CreatedAt)
		return err
	})
	if err != nil {
		log.Printf("insert pings row: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, &p)
}

func (a *MyApp) handleGetPingLatest(c echo.Context) error {
	ctx := c.Request().Context()
	var p ping
	err := a.pool.QueryRow(ctx, "SELECT id, created_at FROM pings ORDER BY id DESC LIMIT 1").Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		log.Printf("get latest pings row: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, &p)
}
