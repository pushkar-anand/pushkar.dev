package main

import (
	"embed"
	"html/template"
	"log"

	"go.uber.org/zap"

	"github.com/pushkar-anand/pushkar.dev/config"
)

//go:embed static
var static embed.FS

//go:embed views/*
var templatesFS embed.FS

func main() {
	// Read app config from environment
	appConfig := config.GetAppConfig()

	// Create logger
	logger := config.NewLogger(appConfig)

	log.Println(appConfig.IsProduction)

	// Read templates
	templates, err := template.ParseFS(templatesFS, "views/*.html")
	if err != nil {
		logger.With(zap.Error(err)).Panic("error loading templates")
	}

	server := NewServer(logger, appConfig, templates)
	server.Initialize()

	server.Listen()

	<-server.connClose
	logger.Info("Shutdown complete")
}
