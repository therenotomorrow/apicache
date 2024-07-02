package test

import (
	"errors"
	"sync"
	"time"
)

// DriverMock implements Driver interface for testing concurrent futures.
type DriverMock struct {
	Storage *sync.Map
	// IsConcurrent uses to make all operations more "slowly"
	// to have more availability with playing concurrent futures.
	IsConcurrent bool
	// Closed emulates the resource closed or not.
	Closed int
}

const (
	// KeyError emulates internal `Storage` errors.
	KeyError = "error"
	// KeyNotExist emulates key not exist behaviour.
	KeyNotExist = "notexist"
	// InternalError emulates internal `Storage` error message.
	InternalError = "internal error"
	// concurrentTimeout if `IsConcurrent == true`.
	concurrentTimeout = 2
)

// Get emulates behaviour: gets key from key-value storage.
func (d *DriverMock) Get(key string) (string, error) {
	if d.IsConcurrent {
		time.Sleep(concurrentTimeout * time.Second)
	}

	if key == KeyError {
		return "", errors.New(InternalError)
	}

	if key == KeyNotExist {
		return "", nil
	}

	val, _ := d.Storage.Load(key)

	return val.(string), nil
}

// Set emulates behaviour: sets key, value and "time-to-live" to key-value storage.
func (d *DriverMock) Set(key, val string, ttl int) error {
	if d.IsConcurrent {
		time.Sleep(concurrentTimeout * time.Second)
	}

	if key == KeyError {
		return errors.New(InternalError)
	}

	if ttl < 0 {
		d.Storage.Delete(key)
	}

	d.Storage.Store(key, val)

	return nil
}

// Delete emulates behaviour: deletes key from key-value storage.
func (d *DriverMock) Delete(key string) (bool, error) {
	if d.IsConcurrent {
		time.Sleep(concurrentTimeout * time.Second)
	}

	if key == KeyError {
		return false, errors.New(InternalError)
	}

	_, ok := d.Storage.Load(key)

	return ok, nil
}

// Close emulates behavior: calls to release key-value storage resources.
func (d *DriverMock) Close() {
	if d.IsConcurrent {
		time.Sleep(concurrentTimeout * time.Second)
	}

	d.Closed = 1
}
