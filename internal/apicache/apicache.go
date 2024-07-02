package apicache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kxnes/go-interviews/apicache/internal/fs"
)

type (
	// Options contains `Server` specific parameters, like `Addr`.
	Options struct {
		Addr string `json:"addr"`
	}
	// Dependencies represents external dependencies that `Server` has.
	Dependencies struct {
		Driver fs.Driver
	}
	// Server represents the main APICache Server.
	Server struct {
		http.Server
		deps *Dependencies
		opts *Options
		done chan struct{}
	}
	// MarshalError decorates outgoing responses to for marshalling `error` type.
	MarshalError struct {
		error
	}
)

// MarshalJSON implementation for `MarshallError` decorator.
func (e *MarshalError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Error())
}

// stop provides all `Shutdown()`-specific dependencies, like free resources.
func (srv *Server) stop() {
	defer func() {
		srv.deps.Driver.Close()
		close(srv.done)
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	<-sigint

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("server shutdown err = %v", err)
	}
}

// routing builds inner `Server` routing.
func (srv *Server) routing() {
	mux := http.NewServeMux()
	mux.Handle("/", &StorageHandler{driver: srv.deps.Driver})
	srv.Handler = mux
}

// Listen aggregates `ListenAndServe()` and adds signal listener for graceful shutdown.
func (srv *Server) Listen() {
	go srv.stop()

	log.Printf("server listen on http://%s (%d)", srv.opts.Addr, syscall.Getpid())

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Panicf("server listen err = %v", err)
	}

	<-srv.done
}

// NewServer returns new `Server`.
func NewServer(deps *Dependencies, opts *Options) *Server {
	srv := &Server{
		Server: http.Server{Addr: opts.Addr},
		deps:   deps,
		opts:   opts,
		done:   make(chan struct{}),
	}

	srv.routing()

	return srv
}

type (
	// StorageHandler handles all specific routes for interrupt with inner `fs.Driver`.
	StorageHandler struct {
		driver fs.Driver
	}
	// request uses for unmarshal incoming POST requests.
	request struct {
		Key string `json:"key"`
		Val string `json:"val"`
		TTL int    `json:"ttl"`
	}
	// Response uses to goal the same interface for all requests.
	Response struct {
		status int
		Val    string `json:"value,omitempty"`
		Err    error  `json:"error,omitempty"`
	}
	// ErrInvalidJSON occurred if incoming POST request cannot parse as JSON.
	ErrInvalidJSON struct{}
	// ErrInvalidType occurred if incoming POST request is correct JSON but missing types.
	ErrInvalidType struct {
		field string
		_type string
	}
)

func (e *ErrInvalidJSON) Error() string {
	return "invalid JSON"
}

func (e *ErrInvalidType) Error() string {
	return fmt.Sprintf("invalid type (%s) for field (%s)", e._type, e.field)
}

// Get contains all GET method logic for `StorageHandler`.
func (api *StorageHandler) Get(r *http.Request) *Response {
	var (
		etc  *fs.ErrConcurrentTimeout
		ene  *fs.ErrNotExist
		resp = new(Response)
	)

	key := r.URL.EscapedPath()[1:]
	val, err := api.driver.Get(key)

	if err != nil {
		resp.Err = err
		wrapped := errors.Unwrap(err)

		switch {
		case wrapped != nil:
			resp.status = http.StatusInternalServerError
			resp.Err = wrapped
		case errors.As(err, &etc):
			resp.status = http.StatusRequestTimeout
		case errors.As(err, &ene):
			resp.status = http.StatusNotFound
		default:
			resp.status = http.StatusBadRequest
		}
	} else {
		resp.status = http.StatusOK
		resp.Val = val
	}

	return resp
}

// Post contains all POST method logic for `StorageHandler`.
func (api *StorageHandler) Post(r *http.Request) *Response {
	var (
		req  request
		etc  *fs.ErrConcurrentTimeout
		ute  *json.UnmarshalTypeError
		resp = new(Response)
	)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		resp.status = http.StatusBadRequest

		switch {
		case errors.As(err, &ute):
			resp.Err = &ErrInvalidType{
				field: ute.Field,
				_type: ute.Type.String(),
			}
		default:
			resp.Err = &ErrInvalidJSON{}
		}

		return resp
	}

	defer func() { _ = r.Body.Close() }()

	err = api.driver.Set(req.Key, req.Val, req.TTL)

	if err != nil {
		resp.Err = err
		wrapped := errors.Unwrap(err)

		switch {
		case wrapped != nil:
			resp.status = http.StatusInternalServerError
			resp.Err = wrapped
		case errors.As(err, &etc):
			resp.status = http.StatusRequestTimeout
		default:
			resp.status = http.StatusBadRequest
		}
	} else {
		resp.status = http.StatusCreated
	}

	return resp
}

// Delete contains all DELETE method logic for `StorageHandler`.
func (api *StorageHandler) Delete(r *http.Request) *Response {
	var (
		etc  *fs.ErrConcurrentTimeout
		ene  *fs.ErrNotExist
		resp = new(Response)
	)

	key := r.URL.EscapedPath()[1:]
	_, err := api.driver.Delete(key)

	if err != nil {
		resp.Err = err
		wrapped := errors.Unwrap(err)

		switch {
		case wrapped != nil:
			resp.status = http.StatusInternalServerError
			resp.Err = wrapped
		case errors.As(err, &etc):
			resp.status = http.StatusRequestTimeout
		case errors.As(err, &ene):
			resp.status = http.StatusNotFound
		default:
			resp.status = http.StatusBadRequest
		}
	} else {
		resp.status = http.StatusNoContent
	}

	return resp
}

func (api *StorageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var resp *Response

	switch r.Method {
	case http.MethodGet:
		resp = api.Get(r)
	case http.MethodPost:
		resp = api.Post(r)
	case http.MethodDelete:
		resp = api.Delete(r)
	default:
		resp = &Response{
			status: http.StatusMethodNotAllowed,
		}
	}

	if resp.Err != nil {
		resp.Err = &MarshalError{resp.Err}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.status)

	if resp.status == http.StatusCreated || resp.status == http.StatusNoContent {
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("response write err = %v\n", err)
	}
}
