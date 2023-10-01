package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukasschwab/tiir/pkg/text"

	_ "github.com/libsql/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

// SQL client defaults.
const (
	defaultPingTimeout      = 1 * time.Second
	defaultOperationTimeout = 3 * time.Second
)

// SQL statements to prepare. NOTE: return field order may be signficant; keep
// in sync with scanText.
const (
	initTableQuery = `
	CREATE TABLE IF NOT EXISTS texts (
		id varchar(8) NOT NULL UNIQUE,
		title text NOT NULL,
		url text NOT NULL,
		author text NOT NULL,
		note text NOT NULL,
		timestamp DATETIME NOT NULL
	);
	`
	deleteQuery = `
	DELETE
	FROM texts WHERE id = :id
	RETURNING id, title, url, author, note, timestamp;
	`
	readQuery = `
	SELECT id, title, url, author, note, timestamp
	FROM texts WHERE id = :id;
	`
	upsertQuery = `
	REPLACE INTO texts (id, title, url, author, note, timestamp) 
	VALUES (:id, :title, :url, :author, :note, :timestamp) 
	RETURNING id, title, url, author, note, timestamp;
	`
	listQuery = `
	SELECT id, title, url, author, note, timestamp FROM texts;
	`
)

func UseLibSQL(connectionString string) (Interface, error) {
	return useLibSQL(connectionString)
}

func useLibSQL(connectionString string) (*SQL, error) {
	db, err := sql.Open("libsql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("error opening DB connection: %w", err)
	}
	s := &SQL{
		DB:               db,
		pingTimeout:      defaultPingTimeout,
		operationTimeout: defaultOperationTimeout,
	}
	if err := s.ping(); err != nil {
		return nil, err
	} else if err := s.prepare(); err != nil {
		return nil, err
	}
	return s, nil
}

// SQL implements [Interface] for libSQL; see [UseSql].
type SQL struct {
	*sql.DB

	upsert *sql.Stmt
	read   *sql.Stmt
	list   *sql.Stmt
	delete *sql.Stmt

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

func (s *SQL) prepare() error {
	ctx, cancel := s.operationContext()
	defer cancel()

	if _, err := s.ExecContext(ctx, initTableQuery); err != nil {
		return fmt.Errorf("error creating table: %w", err)
	} else if s.upsert, err = s.Prepare(upsertQuery); err != nil {
		return fmt.Errorf("erorr preparing upsert: %w", err)
	} else if s.read, err = s.Prepare(readQuery); err != nil {
		return fmt.Errorf("error preparing read: %w", err)
	} else if s.list, err = s.Prepare(listQuery); err != nil {
		return fmt.Errorf("erorr preparing list: %w", err)
	} else if s.delete, err = s.Prepare(deleteQuery); err != nil {
		return fmt.Errorf("error preparing delete: %w", err)
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

	t, err := scan(s.delete.QueryRowContext(ctx, sql.Named("id", id)))
	if err != nil {
		return nil, fmt.Errorf("error deleting row: %w", err)
	}

	return t, nil
}

// List implements [Interface].
func (s *SQL) List(c text.Comparator, d text.Direction) ([]*text.Text, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	// NOTE: ideally c, d are expressible in query (ORDER BY) but that only
	// works for *some* comparators.
	rows, err := s.list.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error reading rows: %w", err)
	}
	defer rows.Close()

	var texts []*text.Text
	for rows.Next() {
		t, err := scan(rows)
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

	t, err := scan(s.read.QueryRowContext(ctx, sql.Named("id", id)))
	if err != nil {
		return nil, fmt.Errorf("error loading row: %w", err)
	}

	return t, nil
}

// Upsert implements [Interface].
func (s *SQL) Upsert(t *text.Text) (*text.Text, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	result, err := scan(s.upsert.QueryRowContext(ctx, asNamedArgs(t)...))
	if err != nil {
		return nil, fmt.Errorf("error upserting text: %w", err)
	}

	return result, nil
}

// scannable describes sql.Row and sql.Rows.
type scannable interface {
	Scan(dest ...any) error
}

// NOTE: scan may need to correspond to field order in prepared queries.
func scan(headRow scannable) (*text.Text, error) {
	var t text.Text
	if err := headRow.Scan(&t.ID, &t.Title, &t.URL, &t.Author, &t.Note, &t.Timestamp); err != nil {
		return nil, fmt.Errorf("error scanning text: %w", err)
	}
	return &t, nil
}

func asNamedArgs(t *text.Text) []any {
	return []any{
		sql.Named("id", t.ID),
		sql.Named("title", t.Title),
		sql.Named("url", t.URL),
		sql.Named("author", t.Author),
		sql.Named("note", t.Note),
		sql.Named("timestamp", t.Timestamp),
	}
}
