package main

import (
	"context"
	"log"

	"go.uber.org/zap"
)

const (
	DatabaseURL = "postgres://gopher:P%40ssw0rd@localhost:5432/app?sslmode=disable"
)

func main() {
	// Предустановленный конфиг. Можно выбрать NewProduction/NewDevelopment/NewExample или создать свой
	// Production - уровень логгирования InfoLevel, формат вывода: json
	// Development - уровень логгирования DebugLevel, формат вывода: console
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = logger.Sync() }()
	// можно установить глобальный логгер (но лучше не надо: используйте внедрение зависимостей где это возможно)
	// undo := zap.ReplaceGlobals(logger)
	// defer undo()
	//
	// zap.L().Info("replaced zap's global loggers")
	a := app{}
	if err := a.Init(context.Background(), logger); err != nil {
		log.Fatal(err)
	}
	if err := a.Serve(); err != nil {
		log.Fatal(err)
	}
}
