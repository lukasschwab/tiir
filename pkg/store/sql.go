package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukasschwab/tiir/pkg/text"
)

const (
	defaultPingTimeout      = 1 * time.Second
	defaultOperationTimeout = 3 * time.Second
)

func UseLibSQL(db *sql.DB) (Interface, error) {
	return useLibSQL(db)
}

func useLibSQL(db *sql.DB) (*SQL, error) {
	s := &SQL{
		DB:               db,
		pingTimeout:      defaultPingTimeout,
		operationTimeout: defaultOperationTimeout,
	}
	if err := s.ping(); err != nil {
		return nil, err
	} else if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

// TODO: implement an "if not exists" init step.

// SQL implements [Interface]; see [UseSql].
type SQL struct {
	*sql.DB
	pingTimeout      time.Duration
	operationTimeout time.Duration
}

func (s *SQL) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.pingTimeout)
	defer cancel()

	if err := s.PingContext(ctx); err != nil {
		return fmt.Errorf("unreachable DB: %w", err)
	}

	return nil
}

func (s *SQL) init() error {
	ctx, cancel := s.operationContext()
	defer cancel()

	q := `
	CREATE TABLE IF NOT EXISTS texts (
		id varchar( 8 ) NOT NULL UNIQUE,
		title text NOT NULL,
		url text NOT NULL,
		author text NOT NULL,
		note text NOT NULL,
		timestamp DATETIME NOT NULL
	)
	`
	if _, err := s.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("error creating table: %w", err)
	}

	return nil
}

func (s *SQL) operationContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.operationTimeout)
}

// Delete implements [Interface].
func (s *SQL) Delete(id string) (*text.Text, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	q := `DELETE FROM texts WHERE id = :id RETURNING id, title, url, author, note, timestamp`

	t, err := scanText(s.QueryRowContext(ctx, q, sql.Named("id", id)))
	if err != nil {
		// TODO: distinguish between a scan error and an actual delete error.
		return nil, fmt.Errorf("error deleting row: %w", err)
	}

	return t, nil
}

// List implements [Interface].
func (s *SQL) List(c text.Comparator, d text.Direction) ([]*text.Text, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	// NOTE: ideally c, d are expressible in query: ORDER BY... but that only
	// works for *some* comparators.
	rows, err := s.QueryContext(ctx, "SELECT id, title, url, author, note, timestamp FROM texts")
	if err != nil {
		return nil, fmt.Errorf("error reading rows: %w", err)
	}
	defer rows.Close()

	var texts []*text.Text
	for rows.Next() {
		t, err := scanOneText(rows)
		if err != nil {
			return nil, err
		}
		texts = append(texts, t)
	}

	text.Sort(texts).By(c, d)
	return texts, nil
}

// Read implements [Interface].
func (s *SQL) Read(id string) (*text.Text, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	q := `SELECT id, title, url, author, note, timestamp FROM texts AS t WHERE t.id = :id `
	t, err := scanText(s.QueryRowContext(ctx, q, sql.Named("id", id)))
	if err != nil {
		return nil, fmt.Errorf("error loading row: %w", err)
	}
	return t, nil
}

// Upsert implements [Interface].
func (s *SQL) Upsert(t *text.Text) (*text.Text, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	upsertQuery := `
	REPLACE INTO texts	(id, title, url, author, note, timestamp) 
	VALUES 				(:id, :title, :url, :author, :note, :timestamp) 
	RETURNING			id, title, url, author, note, timestamp
	`
	result, err := scanText(s.QueryRowContext(ctx, upsertQuery, asArgs(t)...))
	if err != nil {
		return nil, fmt.Errorf("error upserting text: %w", err)
	}

	return result, nil
}

func (s *SQL) Drop() error {
	ctx, cancel := s.operationContext()
	defer cancel()

	if _, err := s.ExecContext(ctx, `DROP TABLE texts`); err != nil {
		return fmt.Errorf("error dropping table: %w", err)
	}
	return nil
}

// TODO: turn these back into named values!
func asArgs(t *text.Text) []any {
	return []any{
		sql.Named("id", t.ID),
		sql.Named("title", t.Title),
		sql.Named("url", t.URL),
		sql.Named("author", t.Author),
		sql.Named("note", t.Note),
		sql.Named("timestamp", t.Timestamp),
	}
}

func scanOneText(rows *sql.Rows) (*text.Text, error) {
	var t text.Text
	if err := rows.Scan(&t.ID, &t.Title, &t.URL, &t.Author, &t.Note, &t.Timestamp); err != nil {
		return nil, fmt.Errorf("error scanning text: %w", err)
	}
	return &t, nil
}

func scanText(row *sql.Row) (*text.Text, error) {
	var t text.Text
	if err := row.Scan(&t.ID, &t.Title, &t.URL, &t.Author, &t.Note, &t.Timestamp); err != nil {
		return nil, fmt.Errorf("error scanning text: %w", err)
	}
	return &t, nil
}
