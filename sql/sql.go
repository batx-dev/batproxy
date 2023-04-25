package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB

	// Datasource name.
	DSN string

	Logger logr.Logger

	// Returns the current time. Defaults to time.Now().
	// Can be mocked for tests.
	Now func() time.Time
}

func NewDB(dsn string) *DB {
	db := &DB{
		DSN:    dsn,
		Logger: logr.Discard(),
		Now:    time.Now,
	}

	return db
}

func (db *DB) Open() (err error) {
	if db.DSN == "" {
		return fmt.Errorf("dsn required")
	}

	dve := "sqlite3"
	dsn := ""
	ss := strings.SplitN(db.DSN, "://", 2)

	switch len(ss) {
	case 1:
		dsn = ss[0]
	case 2:
		dve = ss[0]
		dsn = ss[1]
	default:
		return fmt.Errorf("invalid dsn: %s", db.DSN)
	}

	if db.db, err = sql.Open(dve, dsn); err != nil {
		return err
	}

	if err := db.db.Ping(); err != nil {
		return err
	}

	db.db.SetConnMaxLifetime(5 * time.Minute)

	return nil
}

// Close the database connection.
func (db *DB) Close() error {
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

// BeginTx starts a transaction and returns a wrapper Tx type. This type
// provides a reference to the database and a fixed timestamp at the start of
// the transaction. The timestamp allows us to mock time during tests as well.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Return wrapper Tx that includes the transaction start time.
	return &Tx{
		Tx:  tx,
		db:  db,
		now: db.Now().UTC().Truncate(time.Second),
	}, nil
}

// Tx wraps the SQL Tx object to provide a timestamp at the start of the transaction.
type Tx struct {
	*sql.Tx
	db  *DB
	now time.Time
}

// NullString represents a helper wrapper for string.
type NullString string

// Scan implements the Scanner interface.
func (s *NullString) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	strVal, ok := value.(string)
	if !ok {
		return fmt.Errorf("NullString: cannot scan to string")
	}
	*s = NullString(strVal)
	return nil
}

// Value implements the driver Valuer interface.
func (s *NullString) Value() (driver.Value, error) {
	// if nil or empty string
	if s == nil || len(*s) == 0 {
		return nil, nil
	}
	return string(*s), nil
}

const (
	// Datetime represents mysql DATETIME.
	Datetime = "2006-01-02 15:04:05"
)

// UTCTime represents a helper wrapper for time.Time. It automatically converts
// time fields to/from MySQL datetime as UTC zone. Also supports NULL for zero time.
type UTCTime time.Time

// Scan reads a time value from the database.
func (n *UTCTime) Scan(value interface{}) error {
	if value == nil {
		*(*time.Time)(n) = time.Time{}
		return nil
	} else if value, ok := value.([]byte); ok {
		*(*time.Time)(n), _ = time.ParseInLocation(Datetime, string(value), time.UTC)
		return nil
	}
	return fmt.Errorf("UTCTime: cannot scan to time.Time: %T", value)
}

// Value formats a time value for the database.
func (n *UTCTime) Value() (driver.Value, error) {
	if n == nil || (*time.Time)(n).IsZero() {
		return nil, nil
	}
	return (*time.Time)(n).In(time.UTC).Format(Datetime), nil
}

var (
	// CST represents China Standard Time.
	CST = time.FixedZone("CST", 8*60*60)
)

// NullTime represents a helper wrapper for time.Time. It automatically converts
// time fields to/from RFC 3339 format. Also supports NULL for zero time.
type NullTime time.Time

// Scan reads a time value from the database.
func (t *NullTime) Scan(value interface{}) error {
	if value == nil {
		*t = NullTime(time.Time{})
		return nil
	}

	timeVal, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("NullTime: cannot scan to time.Time")
	}
	*t = NullTime(timeVal)
	return nil
}

// Value formats a time value for the database.
func (t *NullTime) Value() (driver.Value, error) {
	if t == nil || (*time.Time)(t).IsZero() { // if nil or empty string
		return nil, nil
	}
	return (*time.Time)(t).In(CST).Format(Datetime), nil
}

// FormatLimitOffset returns a SQL string for a given limit & offset.
// Clauses are only added if limit and/or offset are greater than zero.
func FormatLimitOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf(`LIMIT %d OFFSET %d`, limit, offset)
	} else if limit > 0 {
		return fmt.Sprintf(`LIMIT %d`, limit)
	} else if offset > 0 {
		return fmt.Sprintf(`OFFSET %d`, offset)
	}
	return ""
}
