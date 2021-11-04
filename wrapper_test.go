package txwrapper

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func Test_Transaction_commit(t *testing.T) {
	email := "info@example.com"

	db, closedb, err := createDB()
	if err != nil {
		t.Fatalf("failed to create db, %s", err)
	}
	defer closedb()

	ctx := context.Background()
	if err := createEmailTable(ctx, db); err != nil {
		t.Fatalf("failed to create table, %s", err)
	}

	wrapper := New(db)
	err = wrapper.Transaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		if err := addEmail(ctx, tx, email); err != nil {
			return err
		}

		if err := addEmail(ctx, tx, email); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		t.Errorf("failed to execute transaction, %s", err)
	}

	actual, err := getEmail(email, db)
	if err != nil {
		t.Errorf("failed to get email, %s", err)
	}

	if actual != 2 {
		t.Errorf("expected 2, get: %d", actual)
	}
}

func Test_Transaction_rollback(t *testing.T) {
	email := "info@example.com"

	db, closedb, err := createDB()
	if err != nil {
		t.Fatalf("failed to create db, %s", err)
	}
	defer closedb()

	ctx := context.Background()
	if err := createEmailTable(ctx, db); err != nil {
		t.Fatalf("failed to create table, %s", err)
	}

	wrapper := New(db)
	err = wrapper.Transaction(context.Background(), nil, func(ctx context.Context, tx *sql.Tx) error {
		if err := addEmail(ctx, tx, email); err != nil {
			return err
		}

		if err := addEmail(ctx, tx, email); err != nil {
			return err
		}

		return fmt.Errorf("fake error")
	})
	if err == nil {
		t.Errorf("test fails because there must be an error")
	}

	actual, err := getEmail(email, db)
	if err != nil {
		t.Errorf("failed to get email, %s", err)
	}

	if actual != 0 {
		t.Errorf("expected 0, get: %d", actual)
	}
}

func createDB() (*sql.DB, func() error, error) {
	file, err := ioutil.TempFile("", "db")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create tmp file, %w", err)
	}

	filename := file.Name()
	if err := file.Close(); err != nil {
		return nil, nil, fmt.Errorf("failed to close file, %w", err)
	}

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open db, %w", err)
	}

	closeFn := func() error {
		errDB := db.Close()
		if err := os.Remove(filename); err != nil || errDB != nil {
			return fmt.Errorf(
				"failed to close db or remove temp file, %s, %s", errDB, err,
			)
		}

		return nil
	}

	return db, closeFn, nil
}

func createEmailTable(ctx context.Context, db *sql.DB) error {
	q := `CREATE TABLE IF NOT EXISTS email (
        email VARCHAR(16) NOT NULL
    )`

	if _, err := db.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("failed to execute query, %w", err)
	}

	return nil
}

func addEmail(ctx context.Context, tx *sql.Tx, email string) error {
	q := `insert into email (email) values (?)`

	if _, err := tx.ExecContext(ctx, q, email); err != nil {
		return fmt.Errorf("failed to execute query, %w", err)
	}

	return nil
}

func getEmail(email string, db *sql.DB) (int, error) {
	var num int
	err := db.QueryRow("select count(*) from email where email=?", email).Scan(&num)
	if err != nil {
		return 0, fmt.Errorf("failed to scan email, %w", err)
	}

	return num, nil
}
