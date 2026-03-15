package database

// import (
// 	"database/sql"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"runtime"
// )

// func migrateTables(db *sql.DB) error {
// 	_, filePath, _, ok := runtime.Caller(0)
// 	if !ok {
// 		return fmt.Errorf("detect current file path")
// 	}

// 	schemaPath := filepath.Join(filepath.Dir(filePath), "schema.sql")
// 	schemaSQL, err := os.ReadFile(schemaPath)
// 	if err != nil {
// 		return fmt.Errorf("read schema file: %w", err)
// 	}

// 	if _, err := db.Exec(string(schemaSQL)); err != nil {
// 		return fmt.Errorf("apply schema: %w", err)
// 	}

// 	return nil
// }
