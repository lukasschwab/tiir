package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lukasschwab/tiir/pkg/text"
)

const (
	defaultPingTimeout      = 1 * time.Second
	defaultOperationTimeout = 3 * time.Second
)

// TODO: this should be UseMySQL.

func UseSQL(db *sql.DB) (Interface, error) {
	s := SQL{
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

func (s SQL) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.pingTimeout)
	defer cancel()

	if err := s.PingContext(ctx); err != nil {
		return fmt.Errorf("unreachable DB: %w", err)
	}

	return nil
}

func (s SQL) init() error {
	ctx, cancel := s.operationContext()
	defer cancel()

	q := `
	CREATE TABLE IF NOT EXISTS texts (
		id varchar( 8 ) NOT NULL UNIQUE,
		title text NOT NULL,
		url text NOT NULL,
		author text NOT NULL,
		note text NOT NULL,
		timestamp timestamp NOT NULL
	)
	`
	if _, err := s.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("error creating table: %w", err)
	}

	return nil
}

func (s SQL) operationContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.operationTimeout)
}

// Delete implements [Interface].
func (s SQL) Delete(id string) (*text.Text, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	q := `
	DELETE FROM texts WHERE id = :id: RETURNING id, title, url, author, note, timestamp
	`

	t, err := scanText(s.QueryRowContext(ctx, q, sql.Named("id", id)))
	if err != nil {
		// TODO: distinguish between a scan error and an actual delete error.
		return nil, fmt.Errorf("error deleting row: %w", err)
	}

	return t, nil
}

// List implements [Interface].
func (s SQL) List(c text.Comparator, d text.Direction) ([]*text.Text, error) {
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

// TODO: standardize query text handling. Also, consider transactionalized inner
// implementations so other functions can call Read.
const ReadQuery = `SELECT id, title, url, author, note, timestamp FROM texts AS t WHERE t.id = ?`

// Read implements [Interface].
func (s SQL) Read(id string) (*text.Text, error) {
	ctx, cancel := s.operationContext()
	defer cancel()
	t, err := scanText(s.QueryRowContext(ctx, ReadQuery, id))
	if err != nil {
		return nil, fmt.Errorf("error loading row: %w", err)
	}
	return t, nil
}

// Upsert implements [Interface].
func (s SQL) Upsert(t *text.Text) (*text.Text, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	tx, err := s.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("can't begin transaction: %w", err)
	}
	defer tx.Rollback()

	// NOTE: these are MYSQL-specific prepared statements.
	upsertQuery := `REPLACE INTO texts (id, title, url, author, note, timestamp) VALUES (?, ?, ?, ?, ?, ?)`
	if res, err := tx.ExecContext(ctx, upsertQuery, namedArgs(t)...); err != nil {
		return nil, fmt.Errorf("error upserting text: %w", err)
	} else if rowsAffected, err := res.RowsAffected(); err != nil {
		return nil, fmt.Errorf("error checking rows affected: %w", err)
	} else {
		switch rowsAffected {
		case 0:
			log.Printf("Careful: upsert affected 0 rows")
		case 1:
			log.Printf("Upsert inserted new text %v", t.ID)
		case 2:
			log.Printf("Upsert updated existing text %v", t.ID)
		}
	}

	// NOTE: this re-read should really be superfluous.
	// if t, err = scanText(tx.QueryRowContext(ctx, ReadQuery, t.ID)); err != nil {
	// 	return nil, fmt.Errorf("error reading text after upsert: %w", err)
	// } else if err := tx.Commit(); err != nil {
	// 	return nil, fmt.Errorf("couldn't commit transaction: %w", err)
	// }

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("couldn't commit transaction: %w", err)
	}

	return t, nil
}

func namedArgs(t *text.Text) []any {
	return []any{
		t.ID,
		t.Title,
		t.URL,
		t.Author,
		t.Note,
		t.Timestamp,
		// sql.Named("id", t.ID),
		// sql.Named("title", t.Title),
		// sql.Named("url", t.URL),
		// sql.Named("author", t.Author),
		// sql.Named("note", t.Note),
		// sql.Named("timestamp", t.Timestamp),
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
