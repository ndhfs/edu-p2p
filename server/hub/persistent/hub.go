package persistent

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"p2p/server/hub"
)

type Hub[T any] struct {
	hub.Hub[T]
	db *sql.DB
}

func NewHub[T any](hub hub.Hub[T]) (*Hub[T], error) {
	db, err := sql.Open("mysql", "p2p:p2p@tcp(127.0.0.1:3325)/p2p?parseTime=true")
	if err != nil {
		return nil, fmt.Errorf("error connect to DB. %w", err)
	}
	return &Hub[T]{Hub: hub, db: db}, nil
}

func (h *Hub[T]) AddClient(ctx context.Context, c hub.Client[T]) error {
	if err := h.Hub.AddClient(ctx, c); err != nil {
		return err
	}

	_, err := h.db.ExecContext(ctx, "INSERT INTO peers (id, name, addr) VALUES (?, ?, ?)", c.Id(), c.Name(), c.Addr())
	if err != nil {
		return fmt.Errorf("error save peer to DB. %w", err)
	}

	return nil
}

// RemoveClient удаляет клиента из хаба
func (h *Hub[T]) RemoveClient(ctx context.Context, c hub.Client[T]) error {
	if err := h.Hub.RemoveClient(ctx, c); err != nil {
		return err
	}

	_, err := h.db.ExecContext(ctx, "DELETE FROM peers WHERE id = ?", c.Id())
	if err != nil {
		return fmt.Errorf("error delete peer from DB. %w", err)
	}

	return nil
}
