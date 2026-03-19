package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"tz/internal/domain"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dsn string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	r := &SQLiteRepository{db: db}
	if err := r.migrate(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *SQLiteRepository) migrate() error {
	_, err := r.db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE IF NOT EXISTS shipments (
			id               TEXT PRIMARY KEY,
			reference_number TEXT NOT NULL UNIQUE,
			origin           TEXT NOT NULL,
			destination      TEXT NOT NULL,
			current_status   TEXT NOT NULL,
			driver_details   TEXT NOT NULL,
			amount           REAL NOT NULL,
			driver_revenue   REAL NOT NULL,
			created_at       TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS events (
			id          TEXT PRIMARY KEY,
			shipment_id TEXT NOT NULL,
			status      TEXT NOT NULL,
			message     TEXT,
			created_at  TEXT NOT NULL,
			FOREIGN KEY (shipment_id) REFERENCES shipments(id)
		);
	`)
	return err
}

func (r *SQLiteRepository) Create(ctx context.Context, s *domain.Shipment, initialEvent *domain.Event) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if initialEvent == nil {
		return fmt.Errorf("%w: initial event is required", domain.ErrInvalidShipment)
	}

	if initialEvent.ID == "" {
		initialEvent.ID = uuid.NewString()
	}
	if initialEvent.CreatedAt.IsZero() {
		initialEvent.CreatedAt = time.Now()
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO shipments
		(id, reference_number, origin, destination, current_status, driver_details, amount, driver_revenue, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.ReferenceNumber, s.Origin, s.Destination,
		string(s.CurrentStatus), s.DriverDetails,
		s.Amount, s.DriverRevenue,
		s.CreatedAt.Format(time.RFC3339),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO events (id, shipment_id, status, message, created_at) VALUES (?, ?, ?, ?, ?)`,
		initialEvent.ID, initialEvent.ShipmentID, string(initialEvent.Status), initialEvent.Message,
		initialEvent.CreatedAt.Format(time.RFC3339),
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id string) (*domain.Shipment, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, reference_number, origin, destination, current_status, driver_details, amount, driver_revenue, created_at
		FROM shipments WHERE id = ?`, id)

	var s domain.Shipment
	var createdAt string
	err := row.Scan(
		&s.ID, &s.ReferenceNumber, &s.Origin, &s.Destination,
		&s.CurrentStatus, &s.DriverDetails,
		&s.Amount, &s.DriverRevenue, &createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrShipmentNotFound
		}
		return nil, err
	}
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &s, nil
}

func (r *SQLiteRepository) UpdateStatusWithEvent(ctx context.Context, id string, status domain.Status, e *domain.Event) error {
	if e == nil {
		return fmt.Errorf("%w: status event is required", domain.ErrInvalidShipment)
	}

	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`UPDATE shipments SET current_status = ? WHERE id = ?`,
		string(status), id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrShipmentNotFound
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO events (id, shipment_id, status, message, created_at) VALUES (?, ?, ?, ?, ?)`,
		e.ID, e.ShipmentID, string(e.Status), e.Message,
		e.CreatedAt.Format(time.RFC3339),
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteRepository) GetHistory(ctx context.Context, shipmentID string) ([]*domain.Event, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, shipment_id, status, message, created_at
		FROM events WHERE shipment_id = ? ORDER BY created_at ASC`, shipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		var e domain.Event
		var createdAt string
		if err := rows.Scan(&e.ID, &e.ShipmentID, &e.Status, &e.Message, &createdAt); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		events = append(events, &e)
	}
	return events, rows.Err()
}
