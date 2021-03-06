package handlers

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pushkar-anand/pushkar.dev/request"

	"go.uber.org/zap"
)

type timer interface {
	// Now returns the current time
	Now() time.Time
	// Since returns the time passed since the given time
	Since(time.Time) time.Duration
}

// realClock save request times
type realClock struct{}

// Now wraps time.Now() to return current time
func (rc *realClock) Now() time.Time {
	return time.Now()
}

// Since returns the duration since the given time
func (rc *realClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

// LoggingHandler is a wrapper to store the logger
type LoggingHandler struct {
	logger  *zap.Logger
	handler http.Handler
	clock   timer
}

// NewLoggingHandler creates a new instance of LoggingHandler.
func NewLoggingHandler(logger *zap.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &LoggingHandler{
			logger:  logger,
			handler: h,
			clock:   &realClock{},
		}
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

// WriteHeader writes a status code to http response
func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	return lw.ResponseWriter.Write(b)
}

func (l *LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestStartTime := l.clock.Now()

	logEntry := l.logger.With(
		zap.String("IP", l.getRealIP(r)),
		zap.String("timestamp", requestStartTime.Format("02/Jan/2006:15:04:05 -0700")),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("protocol", r.Proto),
	)

	lw := newLoggingResponseWriter(w)

	l.handler.ServeHTTP(lw, r)

	requestStopTime := l.clock.Since(requestStartTime)
	rID, err := request.GetRequestID(r)

	if err != nil {
		l.logger.With(zap.Error(err)).Error("error reading request-id")
	}

	logEntry.With(
		zap.String("request-id", rID),
		zap.Int("status", lw.statusCode),
		zap.Duration("processing_time", requestStopTime),
	).Debug("HTTP Request received")
}

// getRealIP - returns real IP from http request
func (l *LoggingHandler) getRealIP(req *http.Request) string {
	var ip string

	remoteAddr := req.RemoteAddr

	if ip = req.Header.Get("X-Forwarded-For"); ip != "" {
		remoteAddr = strings.Split(ip, ", ")[0]
	} else if ip = req.Header.Get("X-Real-IP"); ip != "" {
		remoteAddr = ip
	} else {
		var err error
		remoteAddr, _, err = net.SplitHostPort(remoteAddr)
		if err != nil {
			l.logger.Error("error reading IP", zap.Error(err))
		}
	}

	return remoteAddr
}
