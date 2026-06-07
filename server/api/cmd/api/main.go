// API 进程入口。
//
// 这里只负责装配配置、选择运行时 store、创建 HTTP server。
// 具体业务规则应放在 httpapi / store / auth 等内部包里，避免 main 变成业务入口。
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"memotree/server/api/internal/config"
	"memotree/server/api/internal/httpapi"
	"memotree/server/api/internal/store"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// 默认使用内存 store，方便无数据库时快速启动和跑本地闭环。
	// 配置 MYSQL_DSN 后切换到 MySQL store，并在启动时应用本地 schema。
	appStore := store.Store(store.NewMemoryStore())
	var closeStore func() error
	if cfg.MySQLDSN != "" {
		db, mysqlStore, err := store.OpenMySQL(ctx, cfg.MySQLDSN)
		if err != nil {
			log.Fatalf("connect mysql: %v", err)
		}
		closeStore = db.Close
		defer closeStore()

		schemaSQL, err := os.ReadFile(findMigrationPath())
		if err != nil {
			log.Fatalf("read schema migration: %v", err)
		}
		if err := store.ApplySchema(ctx, db, string(schemaSQL)); err != nil {
			log.Fatalf("apply schema migration: %v", err)
		}
		appStore = mysqlStore
		log.Printf("memotree api using mysql store")
	} else {
		log.Printf("memotree api using in-memory store")
	}

	server := &http.Server{
		Addr:    cfg.APIAddr,
		Handler: httpapi.NewRouterWithStore(cfg, appStore),
	}

	log.Printf("memotree api listening on %s", cfg.APIAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("api server stopped: %v", err)
	}
}

// findMigrationPath 兼容两种启动位置：
// - 在 server/ 目录运行 go run ./api/cmd/api
// - 在仓库根目录运行 go run ./server/api/cmd/api
func findMigrationPath() string {
	candidates := []string{
		filepath.Join("migrations", "0001_initial_schema.sql"),
		filepath.Join("server", "migrations", "0001_initial_schema.sql"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return candidates[0]
}
