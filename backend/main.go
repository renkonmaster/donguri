package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ras0q/goalie"
	"github.com/renkonmaster/donguri/infrastructure/config"
	"github.com/renkonmaster/donguri/infrastructure/database"
	"github.com/renkonmaster/donguri/infrastructure/injector"
)

//go:embed db/init.sql/01_schema.sql
var schemaSQL string

// flushingResponseWriter は Write のたびに Flush を呼ぶ ResponseWriter ラッパーです。
// ogen の SSE エンコーダーが io.Copy を使うため Flush が呼ばれず
// クライアントにイベントが届かない問題を回避します。
type flushingResponseWriter struct {
	http.ResponseWriter
	flusher http.Flusher
}

func (w *flushingResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.flusher.Flush()

	return n, err
}

// sseFlushMiddleware は SSE ストリームエンドポイント (/stream) に対して
// 自動 Flush を行うミドルウェアです。
func sseFlushMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if flusher, ok := w.(http.Flusher); ok && strings.HasSuffix(r.URL.Path, "/stream") {
			next.ServeHTTP(&flushingResponseWriter{ResponseWriter: w, flusher: flusher}, r)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("runtime error: %+v", err)
	}
}

func run() (err error) {
	g := goalie.New()
	defer g.Collect(&err)

	var c config.Config
	c.Parse()

	// connect to and migrate database
	db, err := database.SetupGORM(c.PostgreSQLDSN())
	if err != nil {
		return err
	}

	// ---------------------
	// 自動初期化スキームの実行
	// ---------------------
	if err := db.Exec(schemaSQL).Error; err != nil {
		return fmt.Errorf("failed to execute schema.sql: %w", err)
	}
	// ---------------------

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer g.Guard(sqlDB.Close)

	server, err := injector.InjectServer(db)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	// APIリクエストは ogen のサーバーで処理する
	mux.Handle("/api/", server)

	// その他のリクエストはフロントエンドの静的ファイルを返す
	fs := http.FileServer(http.Dir("/app/dist"))
	mux.Handle("/", fs)

	if err := http.ListenAndServe(c.AppAddr, sseFlushMiddleware(mux)); err != nil {
		return err
	}

	return nil
}
