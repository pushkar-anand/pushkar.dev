package request

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// ContextKey is used for context.Context value. The value requires a key that is not primitive type.
type ContextKey string

// contextKeyRequestID is the ContextKey for RequestID
const contextKeyRequestID ContextKey = "requestID"

// ErrorReqIDRetrieval is returned when request ID couldn't be retrieved
var ErrorReqIDRetrieval = errors.New("request ID couldn't be retrieved")

func AddRequestIDToCTX(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, contextKeyRequestID, reqID)
}

func GetRequestID(r *http.Request) (string, error) {
	return GetRequestIDFromCTX(r.Context())
}

func GetRequestIDFromCTX(ctx context.Context) (string, error) {
	reqID := ctx.Value(contextKeyRequestID)
	if ret, ok := reqID.(string); ok {
		return ret, nil
	}

	return "", ErrorReqIDRetrieval
}

const (
	// HTTPHeaderNameRequestID has the name of the header for request ID
	HTTPHeaderNameRequestID = "X-Request-ID"
)

// assignRequestID will attach a brand new request ID to an incoming http request
func assignRequestID(r *http.Request) *http.Request {
	ctx := AddRequestIDToCTX(r.Context(), uuid.New().String())
	return r.WithContext(ctx)
}

// AssignRequestIDHandler is handler to assign request ID to each incoming request
// Make sure this is the last handler to ensure request ID is assigned as early as possible
func AssignRequestIDHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = assignRequestID(r)
		id, _ := GetRequestID(r)
		w.Header().Set(HTTPHeaderNameRequestID, id)
		next.ServeHTTP(w, r)
	})
}
