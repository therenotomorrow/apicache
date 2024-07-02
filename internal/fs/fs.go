package fs

import (
	"fmt"
	"log"
	"time"
)

const (
	// ErrKVStorage aggregates internal `Driver` (storage) errors.
	ErrKVStorage = "storage error: %w"
	queueDelay   = 100
	minInt       = 1
)

type (
	// ErrConcurrentTimeout occurred if `Driver` is busy to accept another connections.
	ErrConcurrentTimeout struct {
		op string
	}
	// ErrCloseDriver occurred if `Driver` resources closed.
	ErrCloseDriver struct{}
	// ErrEmptyKey occurred if key is not passed.
	ErrEmptyKey struct{}
	// ErrEmptyVal occurred if val is not passed.
	ErrEmptyVal struct {
		key string
	}
	// ErrInvalidTTL occurred if ttl thyself or it's validation is not passed.
	ErrInvalidTTL struct {
		key string
		ttl int
	}
	// ErrNotExist occurred if key not found.
	ErrNotExist struct {
		key string
	}
)

func (e *ErrConcurrentTimeout) Error() string {
	return fmt.Sprintf("timeout for operation (%s)", e.op)
}

func (e *ErrCloseDriver) Error() string {
	return "driver is closed"
}

func (e *ErrEmptyKey) Error() string {
	return "empty key"
}

func (e *ErrEmptyVal) Error() string {
	return fmt.Sprintf("empty value for key (%s)", e.key)
}

func (e *ErrInvalidTTL) Error() string {
	return fmt.Sprintf("invalid ttl (%d) for key (%s)", e.ttl, e.key)
}

func (e *ErrNotExist) Error() string {
	return fmt.Sprintf("key (%s) not exist", e.key)
}

const (
	opGet = "get"
	opSet = "set"
	opDel = "del"
)

type (
	// Driver represents a main interface that available
	// for APICache to manipulate with inner key-value storage.
	Driver interface {
		// Get gets key from key-value storage.
		// Calling packages waits that `Get()` will not return error if key not exist.
		// Otherwise `err` will be wrapped in `ErrKVStorage`.
		Get(key string) (val string, err error)
		// Set sets key, value and "time-to-live" to key-value storage.
		Set(key, val string, ttl int) (err error)
		// Delete deletes key from key-value storage.
		Delete(key string) (ok bool, err error)
		// Close calls to release key-value storage resources.
		Close()
	}
	// Options contains `Driver` specific parameters.
	Options struct {
		MaxConn int           `json:"maxConn"`
		Timeout time.Duration `json:"timeout"`
	}
	// fileSystem implements `Driver` interface.
	fileSystem struct {
		driver Driver
		opts   *Options
		done   chan struct{}
		queue  chan struct{}
	}
)

// acquire checks `Driver`'s availability to process incoming calls.
func (d *fileSystem) acquire(op string) error {
	ticker := time.NewTicker(d.opts.Timeout * time.Second)
	defer ticker.Stop()

	select {
	case <-d.done:
		return &ErrCloseDriver{}
	default:
	}

	select {
	case d.queue <- struct{}{}:
		return nil
	case <-ticker.C:
		return &ErrConcurrentTimeout{op}
	}
}

// release releases one call from "connection pool"
// to give availability to process another incoming calls.
func (d *fileSystem) release() {
	<-d.queue
}

// Get gets key from key-value storage.
func (d *fileSystem) Get(key string) (string, error) {
	if key == "" {
		return "", &ErrEmptyKey{}
	}

	if err := d.acquire(opGet); err != nil {
		return "", err
	}
	defer d.release()

	val, err := d.driver.Get(key)
	if err != nil {
		return "", fmt.Errorf(ErrKVStorage, err)
	}

	if val == "" {
		return "", &ErrNotExist{key}
	}

	return val, nil
}

// Set sets key, value and "time-to-live" to key-value storage.
func (d *fileSystem) Set(key, val string, ttl int) error {
	if key == "" {
		return &ErrEmptyKey{}
	}

	if val == "" {
		return &ErrEmptyVal{key}
	}

	if ttl < minInt {
		return &ErrInvalidTTL{key, ttl}
	}

	if err := d.acquire(opSet); err != nil {
		return err
	}
	defer d.release()

	err := d.driver.Set(key, val, ttl)
	if err != nil {
		return fmt.Errorf(ErrKVStorage, err)
	}

	return nil
}

// Delete deletes key from key-value storage.
func (d *fileSystem) Delete(key string) (bool, error) {
	if key == "" {
		return false, &ErrEmptyKey{}
	}

	if err := d.acquire(opDel); err != nil {
		return false, err
	}
	defer d.release()

	ok, err := d.driver.Delete(key)

	if err != nil {
		return false, fmt.Errorf(ErrKVStorage, err)
	}

	if !ok {
		return false, &ErrNotExist{key}
	}

	return true, nil
}

// Close calls to release key-value storage resources.
// After `d` is `done` all processed will be blocking until `release()`.
func (d *fileSystem) Close() {
	defer d.driver.Close()

	close(d.done)

	// waiting (yes, for infinite time if needed)
	for len(d.queue) != 0 {
		time.Sleep(queueDelay * time.Millisecond)
	}
}

// New returns "ready-to-use" `Driver`.
func New(driver Driver, opts *Options) Driver {
	if opts.Timeout < minInt {
		log.Panicf("non-positive Timeout")
	}

	if opts.MaxConn < 0 {
		log.Panicf("non-positive MaxConn")
	}

	return &fileSystem{
		driver: driver,
		opts:   opts,
		done:   make(chan struct{}),
		queue:  make(chan struct{}, opts.MaxConn),
	}
}
