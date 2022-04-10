package main

//prometheus

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/infrastructure/api/handler"
	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/infrastructure/api/routergin"
	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/infrastructure/db/pgstore"
	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/infrastructure/server"
	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/usecase/app/repo"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	//ust := memory.NewLinks()
	//ust, err := userfilemanager.NewUsers("./data.json", "mem://userRefreshTopic")
	lst, err := pgstore.NewLinks(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	us := repo.NewLinks(lst)
	hs := handler.NewHandlers(us)
	// h := defmux.NewRouter(hs)
	h := routergin.NewRouterGin(hs)
	//h := routeropenapi.NewRouterOpenAPI(hs)
	srv := server.NewServer(":"+os.Getenv("PORT"), h)

	a := server.App{}
	if err := a.Init(); err != nil {
		log.WithFields(log.Fields{
			"Init metrics": time.Now(),
		}).Fatal()
	}
	if err := a.Serve(); err != nil {
		log.WithFields(log.Fields{
			"Server start prometheus": time.Now(),
		}).Info()
	}

	srv.Start(us)
	log.WithFields(log.Fields{
		"Start": time.Now(),
	}).Info()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	for {
		select {
		case <-ctx.Done():
			return
		case <-sigCh:
			log.WithFields(log.Fields{
				"SIGINT": <-sigCh,
			}).Info("cencel context")
			srv.Stop()
			cancel() //Если пришёл сигнал SigInt - завершаем контекст
			lst.Close()
		}
	}
}
