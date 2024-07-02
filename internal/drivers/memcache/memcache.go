package memcache

import (
	"errors"

	"github.com/bradfitz/gomemcache/memcache"
)

// Get gets key from key-value storage.
func (r *Driver) Get(key string) (string, error) {
	val, err := r.storage.Get(key)

	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			err = nil
		}

		return "", err
	}

	return string(val.Value), nil
}

// Set sets key, value and "time-to-live" to key-value storage.
// memcached storage deletes `key` if `ttl < 0`. It is expected behavior.
func (r *Driver) Set(key, val string, ttl int) error {
	return r.storage.Set(&memcache.Item{
		Key:        key,
		Value:      []byte(val),
		Expiration: int32(ttl),
	})
}

// Delete deletes key from key-value storage.
func (r *Driver) Delete(key string) (bool, error) {
	err := r.storage.Delete(key)

	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			err = nil
		}

		return false, err
	}

	return true, nil
}

// Close calls to release key-value storage resources.
func (r *Driver) Close() {}

// Driver implements Driver interface.
type Driver struct {
	storage *memcache.Client
}

// New returns "ready-to-use" `Driver` with `memcache` inner storage.
func New(addr string) *Driver {
	return &Driver{
		storage: memcache.New(addr),
	}
}
