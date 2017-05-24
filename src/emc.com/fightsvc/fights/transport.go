package fights

// The fightsvc is just over HTTP, so we just have a single transport.go.

import (
	_ "bytes"
	"context"
	"encoding/json"
	"errors"
	_ "io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
)

var (
	// ErrBadRouting is returned when an expected path variable is missing.
	// It always indicates programmer error.
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
// Useful in a fightsvc server.
func MakeHTTPHandler(s Service, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	// POST    /fights/                          adds another fight
	r.Methods("POST").Path("/fights/").Handler(httptransport.NewServer(
		e.PostFightEndpoint,
		decodePostFightRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/fights/{id}").Handler(httptransport.NewServer(
		e.GetFightEndpoint,
		decodeGetFightRequest,
		encodeResponse,
		options...,
	))

	r.Methods("PUT").Path("/fights/{id}").Handler(httptransport.NewServer(
		e.PutFightEndpoint,
		decodePutFightRequest,
		encodeResponse,
		options...,
	))

	r.Methods("DELETE").Path("/fights/{id}").Handler(httptransport.NewServer(
		e.DeleteFightEndpoint,
		decodeDeleteFightRequest,
		encodeDeleteResponse,
		options...,
	))

	return r
}

func decodePostFightRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req postFightRequest
	if e := json.NewDecoder(r.Body).Decode(&req.Fight); e != nil {
		return nil, e
	}
	return req, nil
}

func decodeGetFightRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}

	req := getFightRequest{ID: id}
	return req, nil
}

func decodePutFightRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}

	var attack map[string]interface{}
	if e := json.NewDecoder(r.Body).Decode(&attack); e != nil {
		return nil, e
	}

	req := putFightRequest{ID: id, Attack: attack}
	return req, nil
}

func decodeDeleteFightRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}

	req := deleteFightRequest{ID: id}
	return req, nil
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error. For more information, read the
// big comment in endpoints.go.
type errorer interface {
	error() error
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func encodeDeleteResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNoContent)
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrAlreadyExists, ErrInconsistentIDs:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
