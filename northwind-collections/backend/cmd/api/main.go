package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpadapter "github.com/ron/northwind-collections/internal/adapters/http"
	"github.com/ron/northwind-collections/internal/adapters/postgres"
	"github.com/ron/northwind-collections/internal/adapters/seed"
	"github.com/ron/northwind-collections/internal/application"
)

func main() {
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := postgres.RunMigrations(ctx, pool, migrationsDir); err != nil {
		log.Fatal("migrations: ", err)
	}

	if os.Getenv("SEED_ON_START") == "true" {
		if err := seed.RunIfEmpty(ctx, pool); err != nil {
			log.Fatal("seed: ", err)
		}
	}

	customerRepo := postgres.NewCustomerRepo(pool)
	invoiceRepo := postgres.NewInvoiceRepo(pool)
	actionRepo := postgres.NewActionRepo(pool)
	dashboardRepo := postgres.NewDashboardRepo(pool)

	svc := application.NewCollectionService(customerRepo, invoiceRepo, actionRepo, dashboardRepo)
	router := httpadapter.NewRouter(svc)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: ":" + port, Handler: router}

	go func() {
		log.Printf("API listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
