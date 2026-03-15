package database

import "database/sql"

// TODO: 初期開発段階のためマイグレーション不要。このファイルごと削除し、
// gorm.go の SetupGORM からの migrateTables 呼び出しも削除すること。
// スキーマ変更時は既存テーブルを DROP して再作成する。
func migrateTables(_ *sql.DB) error {
	return nil
}
