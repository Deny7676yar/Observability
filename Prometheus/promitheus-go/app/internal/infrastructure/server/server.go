package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/usecase/app/repo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	srv http.Server
	ls  *repo.Links
}

func NewServer(addr string, h http.Handler) *Server {
	s := &Server{}

	s.srv = http.Server{
		Addr:              addr,
		Handler:           h,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
	}
	return s
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := s.srv.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{
			"Server": err,
		}).Errorf("Server Shutdown Failed")
	}
	cancel()
}

func (s *Server) Start(ls *repo.Links) {
	s.ls = ls
	var err error
	// TODO: migrations
	go s.srv.ListenAndServe() //nolint
	if err != nil {
		log.WithFields(log.Fields{
			"ListenAndServe": err,
		}).Errorf("ListenAndServe Failed")
	}
}

func (a *App) Serve() error {
	mux := http.NewServeMux()
	mux.Handle("/process", http.HandlerFunc(a.processHandler)) ///process?line=текст+тут
	mux.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe("0.0.0.0:9000", mux)
}
