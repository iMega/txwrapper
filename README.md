# TxWrapper

TxWrapper is a sql transaction wrapper. It helps to exclude writing code for
rollback and commit commands.

### Usage

```go
import (
    "context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"

	_ "github.com/mattn/go-sqlite3"
    "github.com/imega/txwrapper"
)

func main() {
    file, err := ioutil.TempFile("", "db")
	if err != nil {
		log.Fatalf("failed to create tmp file, %w", err)
	}

	filename := file.Name()
	if err := file.Close(); err != nil {
		log.Fatalf("failed to close tmp file, %w", err)
	}

    db, err := sql.Open("sqlite3", filename)
	if err != nil {
        log.Fatalf("failed to open db, %w", err)
	}

    ctx := context.Background()
	w := txwrapper.New(db)
	err := w.Transaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
        if err := createEmailTable(ctx, tx); err != nil {
            return err
        }

        if err := addEmail(ctx, tx, "info@example.com"); err != nil {
            return err
        }

        return nil
    })
    if err != nil {
        log.Fatalf("failed to open db, %w", err)
	}

    errDB := db.Close()
    if err := os.Remove(filename); err != nil || errDB != nil {
        log.Fatalf("failed to close db or remove tmp file, %w, %w", errDB, err)
    }
}

func createEmailTable(ctx context.Context, tx *sql.Tx) error {
	q := `CREATE TABLE IF NOT EXISTS email (
        email VARCHAR(64) NOT NULL
    )`

	if _, err := tx.ExecContext(ctx, q); err != nil {
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
```
