package config

import (
	"os"
	"strconv"
	"sync"
)

// App contains config data for the app
type App struct {
	Environment  Environment
	IsProduction bool
	PORT         int

	SentryDSN string
}

var (
	app        *App
	configOnce sync.Once
)

//GetAppConfig reads and return app config
func GetAppConfig() *App {
	configOnce.Do(func() {
		app = readConfig()
	})

	return app
}

func readConfig() *App {
	portStr := os.Getenv("PORT")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}

	a := &App{
		PORT:         port,
		Environment:  Environment(os.Getenv("ENV")),
		IsProduction: false,
		SentryDSN:    os.Getenv("SENTRY_DSN"),
	}

	if a.Environment == EnvProduction {
		a.IsProduction = true
	}

	return a
}
