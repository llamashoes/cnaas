package fights

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints collects all of the endpoints that compose a fight service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
//
// In a server, it's useful for functions that need to operate on a per-endpoint
// basis. For example, you might pass an Endpoints to a function that produces
// an http.Handler, with each method (endpoint) wired up to a specific path. (It
// is probably a mistake in design to invoke the Service methods on the
// Endpoints struct in a server.)
//
// In a client, it's useful to collect individually constructed endpoints into a
// single type that implements the Service interface. For example, you might
// construct individual endpoints using transport/http.NewClient, combine them
// into an Endpoints, and return it to the caller as a Service.
type Endpoints struct {
	PostFightEndpoint   endpoint.Endpoint
	GetFightEndpoint    endpoint.Endpoint
	PutFightEndpoint    endpoint.Endpoint
	DeleteFightEndpoint endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service. Useful in a fightsvc
// server.
func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		PostFightEndpoint:   MakePostFightEndpoint(s),
		GetFightEndpoint:    MakeGetFightEndpoint(s),
		PutFightEndpoint:    MakePutFightEndpoint(s),
		DeleteFightEndpoint: MakeDeleteFightEndpoint(s),
	}
}

// Primarily useful in a server.
func MakePostFightEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postFightRequest)
		response, err = s.PostFight(ctx, req.Fight)
		return
	}
}

// Primarily useful in a server.
func MakeGetFightEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getFightRequest)
		response, err = s.GetFight(ctx, req.ID)
		return
	}
}

// Primarily useful in a server.
func MakePutFightEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(putFightRequest)
		response, err = s.PutFight(ctx, req.ID, req.Attack)
		return
	}
}

// Primarily useful in a server.
func MakeDeleteFightEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteFightRequest)
		err = s.DeleteFight(ctx, req.ID)
		return
	}
}

// We have two options to return errors from the business logic.
//
// We could return the error via the endpoint itself. That makes certain things
// a little bit easier, like providing non-200 HTTP responses to the client. But
// Go kit assumes that endpoint errors are (or may be treated as)
// transport-domain errors. For example, an endpoint error will count against a
// circuit breaker error count.
//
// Therefore, it's often better to return service (business logic) errors in the
// response object. This means we have to do a bit more work in the HTTP
// response encoder to detect e.g. a not-found error and provide a proper HTTP
// status code. That work is done with the errorer interface, in transport.go.
// Response types that may contain business-logic errors implement that
// interface.

// POST request and response
type postFightRequest struct {
	Fight Fight
}

type postFightResponse struct {
	Fight map[string]interface{}
	Err   error `json:"err,omitempty"`
}

func (r postFightResponse) error() error {
	return r.Err
}

// GET request and response
type getFightRequest struct {
	ID string
}

type getFightResponse struct {
	Fight map[string]interface{}
	Err   error `json:"err,omitempty"`
}

func (r getFightResponse) error() error {
	return r.Err
}

// PUT request and response
type putFightRequest struct {
	ID     string
	Attack map[string]interface{}
}

type putFightResponse struct {
	Fight map[string]interface{}
	Err   error `json:"err,omitempty"`
}

func (r putFightResponse) error() error {
	return r.Err
}

// DELETE
type deleteFightRequest struct {
	ID string
}

type deleteFightResponse struct {
	Err error `json:"err,omitempty"`
}

func (r deleteFightResponse) error() error {
	return r.Err
}
