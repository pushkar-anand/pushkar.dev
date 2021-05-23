package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pushkar-anand/pushkar.dev/security"

	"github.com/pushkar-anand/pushkar.dev/request"

	"github.com/pushkar-anand/pushkar.dev/handlers"

	"github.com/gorilla/mux"
	"github.com/pushkar-anand/pushkar.dev/config"
	"github.com/pushkar-anand/pushkar.dev/template"

	"go.uber.org/zap"

	html "html/template"
)

// Server wraps server data and functions
type Server struct {
	server http.Server

	router      *mux.Router
	HTTPHandler http.Handler
	logger      *zap.Logger
	appConfig   *config.App

	killServer chan int
	connClose  chan int

	renderer *template.Renderer
}

// NewServer creates an instance of the Server
func NewServer(logger *zap.Logger, appConfig *config.App, templates *html.Template) *Server {
	server := &Server{
		logger:     logger,
		appConfig:  appConfig,
		killServer: make(chan int),
		connClose:  make(chan int),
		renderer:   template.NewRenderer(templates, logger),
	}

	r := mux.NewRouter()
	r.StrictSlash(true)

	server.router = r

	return server
}

// Initialize the Server
func (s *Server) Initialize() {
	s.addMiddleWares()
	s.addRoutes()
	s.addHandlers()

	// This runs in background and listens for any signal from the OS
	go func() {
		sigint := make(chan os.Signal)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGKILL)

		select {
		case sig := <-sigint:
			s.logger.Info("Shutdown signal received", zap.String("signal", sig.String()))
		case <-s.killServer:
			s.logger.Info("Received kill server request")
		}

		s.logger.Debug("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := s.server.Shutdown(ctx)
		if err != nil {
			s.logger.With(zap.Error(err)).Error("Error shutting down server")
		}
		close(s.connClose)
	}()

	// To ensure graceful shutdown when a fatal log is logged
	/*s.logger.ExitFunc = func(code int) {
		s.logger.Info("Issuing kill request")
		s.killServer <- 1
	}*/
}

func (s *Server) addRoutes() {
	s.router.PathPrefix("/static/").Handler(http.FileServer(http.FS(static)))

	s.router.HandleFunc("/", func(w http.ResponseWriter, h *http.Request) {
		s.renderer.Render(w, "home.html", nil)
	})
}

func (s *Server) addMiddleWares() {

}

func (s *Server) addHandlers() {
	s.HTTPHandler = s.router

	s.HTTPHandler = handlers.NewLoggingHandler(s.logger)(s.HTTPHandler)
	s.HTTPHandler = security.SecureHandler(s.appConfig.IsProduction)(s.HTTPHandler)
	s.HTTPHandler = handlers.RecoveryHandler(s.logger, s.appConfig.IsProduction)(s.HTTPHandler)
	s.HTTPHandler = request.AssignRequestIDHandler(s.HTTPHandler)
}

// Listen starts the server
func (s *Server) Listen() {
	addr := fmt.Sprintf("%s:%d", "", s.appConfig.PORT)

	s.server = http.Server{
		Addr:         addr,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      s.HTTPHandler,
	}

	s.logger.Info("Server started", zap.String("address", addr))

	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		s.logger.With(zap.Error(err)).Fatal("HTTP server error")
	}
}
