package main

import (
	app "avito.go/internal/app"
	"avito.go/internal/config"
	"avito.go/internal/routes"
	"avito.go/internal/storage"
	"avito.go/pkg/logger"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	err = logger.InitLogger()
	if err != nil {
		fmt.Println("Error init logger ", err)
	}

	DatabaseDSN := storage.GetDatabaseDSN(*cfg)

	fmt.Println(DatabaseDSN)

	db := storage.NewStorage(DatabaseDSN)
	defer db.DB.Close()

	//TODO: покрыть тестами
	//TODO: auth изменить логгирование

	A := app.NewApp(db)
	r := routes.NewRouter(*A)

	srv := http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	done := make(chan struct{})
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt)
		<-sigs
		if err = srv.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(done)
	}()

	err = srv.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}
