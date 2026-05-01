package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
)

func TestOpenInMemory(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if db == nil {
		t.Fatal("expected non-nil DB")
	}
	defer db.Close()
}

func TestOpenAndClose(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := db.Close(); err != nil {
		t.Errorf("expected no error on close, got %v", err)
	}
}

func TestDoubleClose(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	db.Close()
	// Second close should not panic
	_ = db.Close()
}

func TestOpenInvalidPath(t *testing.T) {
	_, err := Open("/nonexistent/directory/db.sqlite")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestConn(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	conn := db.Conn()
	if conn == nil {
		t.Fatal("expected non-nil *sql.DB")
	}
	if err := conn.Ping(); err != nil {
		t.Errorf("expected ping to succeed: %v", err)
	}
}

// ---------- Exec, Query, QueryRow ----------

func TestExec(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	result, err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	_ = result
}

func TestInsertAndQuery(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	db.Exec("CREATE TABLE items (id INTEGER, val TEXT)")
	db.Exec("INSERT INTO items VALUES (1, 'hello')")
	db.Exec("INSERT INTO items VALUES (2, 'world')")

	rows, err := db.Query("SELECT id, val FROM items ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var val string
		if err := rows.Scan(&id, &val); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		count++
		if id == 1 && val != "hello" {
			t.Errorf("expected val=hello for id=1, got %q", val)
		}
		if id == 2 && val != "world" {
			t.Errorf("expected val=world for id=2, got %q", val)
		}
	}
	if count != 2 {
		t.Errorf("expected 2 rows, got %d", count)
	}
}

func TestQueryRow(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	db.Exec("CREATE TABLE test (id INTEGER, val TEXT)")
	db.Exec("INSERT INTO test VALUES (1, 'single')")

	var val string
	err = db.QueryRow("SELECT val FROM test WHERE id = ?", 1).Scan(&val)
	if err != nil {
		t.Fatalf("queryrow failed: %v", err)
	}
	if val != "single" {
		t.Errorf("expected val=single, got %q", val)
	}
}

func TestQueryRowNotFound(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	db.Exec("CREATE TABLE test (id INTEGER)")
	var id int
	err = db.QueryRow("SELECT id FROM test WHERE id = ?", 999).Scan(&id)
	if err == nil {
		t.Error("expected error for not found row")
	}
}

// ---------- Tables ----------

func TestTablesEmpty(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	tables, err := db.Tables()
	if err != nil {
		t.Fatalf("tables() failed: %v", err)
	}
	if len(tables) != 0 {
		t.Errorf("expected 0 tables, got %d: %v", len(tables), tables)
	}
}

func TestTablesWithData(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	db.Exec("CREATE TABLE t1 (id INTEGER)")
	db.Exec("CREATE TABLE t2 (id INTEGER)")

	tables, err := db.Tables()
	if err != nil {
		t.Fatalf("tables() failed: %v", err)
	}
	if len(tables) != 2 {
		t.Errorf("expected 2 tables, got %d: %v", len(tables), tables)
	}
}

func TestTablesExcludesSqliteInternal(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	db.Exec("CREATE TABLE my_table (id INTEGER)")

	tables, err := db.Tables()
	if err != nil {
		t.Fatalf("tables() failed: %v", err)
	}

	// sqlite_sequence may exist if we had AUTOINCREMENT, but shouldn't be returned
	for _, name := range tables {
		if len(name) >= 7 && name[:7] == "sqlite_" {
			t.Errorf("sqlite internal table leaked: %s", name)
		}
	}
}

// ---------- TableExists ----------

func TestTableExists(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	db.Exec("CREATE TABLE my_table (id INTEGER)")

	exists, err := db.TableExists("my_table")
	if err != nil {
		t.Fatalf("tableexists failed: %v", err)
	}
	if !exists {
		t.Error("expected my_table to exist")
	}

	exists, err = db.TableExists("nonexistent")
	if err != nil {
		t.Fatalf("tableexists failed: %v", err)
	}
	if exists {
		t.Error("expected nonexistent to not exist")
	}
}

func TestTableExistsCaseSensitivity(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	db.Exec("CREATE TABLE MyTable (id INTEGER)")

	exists, err := db.TableExists("MyTable")
	if err != nil {
		t.Fatalf("tableexists failed: %v", err)
	}
	if !exists {
		t.Error("expected MyTable to exist")
	}
}

// ---------- Columns ----------

func TestColumns(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT NOT NULL, age INTEGER DEFAULT 0)")

	cols, err := db.Columns("test")
	if err != nil {
		t.Fatalf("columns() failed: %v", err)
	}
	if len(cols) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(cols))
	}

	if cols[0].Name != "id" {
		t.Errorf("expected col[0].name=id, got %q", cols[0].Name)
	}
	if !cols[0].PrimaryKey {
		t.Error("expected id to be primary key")
	}

	if cols[1].Name != "name" {
		t.Errorf("expected col[1].name=name, got %q", cols[1].Name)
	}
	if !cols[1].NotNull {
		t.Error("expected name to be NOT NULL")
	}

	if cols[2].Name != "age" {
		t.Errorf("expected col[2].name=age, got %q", cols[2].Name)
	}
	if !cols[2].Default.Valid {
		t.Error("expected age to have default value")
	}
}

func TestColumnsEmptyTable(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	db.Exec("CREATE TABLE empty (id INTEGER)")

	cols, err := db.Columns("empty")
	if err != nil {
		t.Fatalf("columns() failed: %v", err)
	}
	if len(cols) != 1 {
		t.Errorf("expected 1 column, got %d", len(cols))
	}
}

func TestColumnsNonExistentTable(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	cols, err := db.Columns("no_such_table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cols) != 0 {
		t.Errorf("expected 0 columns for non-existent table, got %d", len(cols))
	}
}

// ---------- Transactions ----------

func TestExecTxCommit(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	err = db.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("CREATE TABLE test (id INTEGER)")
		if err != nil {
			return err
		}
		_, err = tx.Exec("INSERT INTO test VALUES (42)")
		return err
	})
	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}

	// Verify data was committed
	var val int
	db.QueryRow("SELECT id FROM test").Scan(&val)
	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}
}

func TestExecTxRollback(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	err = db.ExecTx(func(tx *sql.Tx) error {
		tx.Exec("CREATE TABLE test (id INTEGER)")
		tx.Exec("INSERT INTO test VALUES (1)")
		return fmt.Errorf("rollback requested")
	})
	if err == nil {
		t.Fatal("expected error from tx")
	}

	// Table should not exist since tx rolled back
	exists, _ := db.TableExists("test")
	if exists {
		t.Error("expected table to not exist after rollback")
	}
}

// ---------- PRAGMAs ----------

func TestWALMode(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	var journalMode string
	db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if journalMode != "wal" && journalMode != "memory" {
		t.Logf("journal_mode=%q (acceptable: wal or memory for :memory:)", journalMode)
	}
}

func TestForeignKeyEnabled(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	var fkEnabled int
	db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	if fkEnabled != 1 {
		t.Errorf("expected foreign_keys=1, got %d", fkEnabled)
	}
}

// ---------- Multiple databases ----------

func TestMultipleDatabaseInstances(t *testing.T) {
	db1, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db1.Close()

	db2, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db2.Close()

	db1.Exec("CREATE TABLE shared (id INTEGER)")
	db1.Exec("INSERT INTO shared VALUES (1)")

	// db2 should not have the table from db1
	exists, _ := db2.TableExists("shared")
	if exists {
		t.Error("db2 should not have db1's table")
	}
}

// ---------- Simultaneous operations ----------

func TestConcurrentQueries(t *testing.T) {
	tmp := t.TempDir() + "/test.db"
	db, err := Open(tmp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer os.Remove(tmp)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE numbers (val INTEGER)")
	if err != nil {
		t.Fatalf("create table failed: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, err := db.Exec("INSERT INTO numbers VALUES (?)", n)
			if err != nil {
				t.Errorf("concurrent insert failed: %v", err)
			}
		}(i)
	}
	wg.Wait()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM numbers").Scan(&count)
	if err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 10 {
		t.Errorf("expected 10 rows, got %d", count)
	}
}

// ---------- Edge cases ----------

func TestExecEmptyQuery(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	// Empty query may or may not error depending on driver; ensure it doesn't panic
	_, _ = db.Exec("")
}
