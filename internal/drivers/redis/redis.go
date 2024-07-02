package redis

import (
	"time"

	"github.com/go-redis/redis/v7"
)

// Get gets key from key-value storage.
func (r *Driver) Get(key string) (string, error) {
	val, err := r.storage.Get(key).Result()

	if err != nil {
		if err == redis.Nil {
			err = nil
		}

		return "", err
	}

	return val, nil
}

// Set sets key, value and "time-to-live" to key-value storage.
func (r *Driver) Set(key, val string, ttl int) error {
	// force `memcache` behaviour because `redis.Set()` ignores negative `ttl`.
	if ttl < 0 {
		r.storage.Del(key)
		return nil
	}

	_, err := r.storage.Set(key, val, time.Duration(ttl)*time.Second).Result()

	return err
}

// Delete deletes key from key-value storage.
func (r *Driver) Delete(key string) (bool, error) {
	val, err := r.storage.Del(key).Result()
	if err != nil {
		return false, err
	}

	if val == 0 {
		return false, nil
	}

	return true, nil
}

// Close calls to release key-value storage resources.
func (r *Driver) Close() {
	_ = r.storage.Close()
}

// Driver implements Driver interface.
type Driver struct {
	storage *redis.Client
}

// New returns "ready-to-use" `Driver` with `redis` inner storage.
func New(addr string) *Driver {
	return &Driver{
		storage: redis.NewClient(&redis.Options{Addr: addr}),
	}
}
