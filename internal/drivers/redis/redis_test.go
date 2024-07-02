package redis

import (
	"errors"
	"log"
	"os"

	"github.com/go-redis/redis/v7"
	"reflect"
	"testing"
	"time"
)

const (
	testInstance = ":6379"

	keyWithoutExpire = "key-without-expire"
	valWithoutExpire = "val-without-expire"
	withoutExpire    = 0 // seconds

	keyWithLongExpire = "key-with-long-expire"
	valWithLongExpire = "val-with-long-expire"
	longExpire        = 100 // seconds

	keyWithShortExpire = "key-with-short-expire"
	valWithShortExpire = "val-with-short-expire"
	shortExpire        = 2 // seconds

	keyNotExist   = "key-not-exist"
	valNotExist   = ""
	invalidExpire = -10 // seconds
)

var d = Driver{storage: redis.NewClient(&redis.Options{})}

func TestMain(m *testing.M) {
	for _, item := range []struct {
		Key, Value string
		Expiration int
	}{
		{
			Key:   keyWithoutExpire,
			Value: valWithoutExpire,
		},
		{
			Key:        keyWithLongExpire,
			Value:      valWithLongExpire,
			Expiration: longExpire,
		},
		{
			Key:        keyWithShortExpire,
			Value:      valWithShortExpire,
			Expiration: shortExpire,
		},
	} {
		err := d.storage.Set(item.Key, item.Value, time.Duration(item.Expiration)*time.Second).Err()
		if err != nil {
			log.Fatalf("setup fixture %v error = %v\n", item, err)
		}
	}

	code := m.Run()

	err := d.storage.FlushAll().Err()
	if err != nil {
		log.Fatalf("teardown error = %v\n", err)
	}

	os.Exit(code)
}

func TestDriverGetKeyExists(t *testing.T) {
	val, err := d.Get(keyWithoutExpire)

	if val != valWithoutExpire {
		t.Errorf("Get() = %s, want = %s", val, valWithoutExpire)
	}

	if err != nil {
		t.Errorf("Get() error = %v, want = %v", err, nil)
	}
}

func TestDriverGetKeyNotExists(t *testing.T) {
	val, err := d.Get(keyNotExist)

	if val != valNotExist {
		t.Errorf("Get() = %s, want = %s", val, valNotExist)
	}

	if err != nil {
		t.Errorf("Get() error = %v, want = %v", err, nil)
	}
}

func TestDriverGetKeyExpired(t *testing.T) {
	time.Sleep(shortExpire * time.Second)

	val, err := d.Get(keyWithShortExpire)

	if val != valNotExist {
		t.Errorf("Get() = %s, want = %s", val, valNotExist)
	}

	if err != nil {
		t.Errorf("Get() error = %v, want = %v", err, nil)
	}
}

func TestDriverGetUnexpectedError(t *testing.T) {
	_ = d.storage.Close()

	val, err := d.Get(valNotExist)

	if val != valNotExist {
		t.Errorf("Get() = %s, want = %s", val, valNotExist)
	}

	if !reflect.DeepEqual(err, errors.New(`redis: client is closed`)) {
		t.Errorf("Get() error = %v, want = %v", err, nil)
	}

	d = *New(testInstance)
}

func TestDriverSetNotExistKey(t *testing.T) {
	item, err := d.storage.Get(keyNotExist).Result()
	if err != redis.Nil || item != "" {
		t.Errorf("Key exists but doesn't")
	}

	err = d.Set(keyNotExist, valWithoutExpire, withoutExpire)

	if err != nil {
		t.Errorf("Set() error = %v, want = %v", err, nil)
	}

	item, err = d.storage.Get(keyNotExist).Result()
	if err == redis.Nil || item == "" {
		t.Errorf("Set() doesn't set")
	}

	if item != valWithoutExpire {
		t.Errorf("Set() wrong value = %s", item)
	}
}

func TestDriverSetExistKey(t *testing.T) {
	item, err := d.storage.Get(keyWithoutExpire).Result()
	if err == redis.Nil || item == "" {
		t.Errorf("Key not exists but doesn't")
	}

	err = d.Set(keyWithoutExpire, valWithoutExpire, withoutExpire)

	if err != nil {
		t.Errorf("Set() error = %v, want = %v", err, nil)
	}

	item, err = d.storage.Get(keyWithoutExpire).Result()
	if err == redis.Nil || item == "" {
		t.Errorf("Set() doesn't set")
	}

	if item != valWithoutExpire {
		t.Errorf("Set() wrong value = %s", item)
	}
}

func TestDriverSetUpdateExpireOnKey(t *testing.T) {
	err := d.Set(keyWithLongExpire, valWithLongExpire, shortExpire)
	if err != nil {
		t.Errorf("Set() error = %v, want = %v", err, nil)
	}

	time.Sleep(time.Second)

	err = d.Set(keyWithLongExpire, valWithLongExpire, shortExpire)
	if err != nil {
		t.Errorf("Set() error = %v, want = %v", err, nil)
	}

	time.Sleep(time.Second)

	item, err := d.storage.Get(keyWithLongExpire).Result()
	if err == redis.Nil || item == "" {
		t.Errorf("Set() doesn't set")
	}

	if item != valWithLongExpire {
		t.Errorf("Set() wrong value = %s", item)
	}

	time.Sleep(time.Second)

	item, err = d.storage.Get(keyWithLongExpire).Result()
	if err != redis.Nil || item != "" {
		t.Errorf("Set() doesn't update expiration")
	}
}

func TestDriverSetNegativeTTL(t *testing.T) {
	err := d.Set(keyWithoutExpire, valWithoutExpire, invalidExpire)
	if err != nil {
		t.Errorf("Set() error = %v, want = %v", err, nil)
	}

	item, err := d.storage.Get(keyWithoutExpire).Result()
	if err != redis.Nil || item != "" {
		t.Errorf("Set() set but doesn't")
	}
}

func TestDriverSetUnexpectedError(t *testing.T) {
	_ = d.storage.Close()

	err := d.Set(valNotExist, valWithoutExpire, longExpire)

	if !reflect.DeepEqual(err, errors.New(`redis: client is closed`)) {
		t.Errorf("Get() error = %v, want = %v", err, nil)
	}

	d = *New(testInstance)
}

func TestDriverDeleteKeyExists(t *testing.T) {
	_ = d.Set(keyWithoutExpire, valWithoutExpire, withoutExpire)

	ok, err := d.Delete(keyWithoutExpire)
	if !ok {
		t.Errorf("Delete() = %v, want = %v", ok, true)
	}

	if err != nil {
		t.Errorf("Delete() error = %v, want = %v", err, nil)
	}
}

func TestDriverDeleteKeyNotExists(t *testing.T) {
	ok, err := d.Delete(valWithoutExpire)
	if ok {
		t.Errorf("Delete() = %v, want = %v", ok, false)
	}

	if err != nil {
		t.Errorf("Delete() error = %v, want = %v", err, nil)
	}
}

func TestDriverDeleteUnexpectedError(t *testing.T) {
	_ = d.storage.Close()

	ok, err := d.Delete(valWithoutExpire)
	if ok {
		t.Errorf("Delete() = %v, want = %v", ok, false)
	}

	if !reflect.DeepEqual(err, errors.New(`redis: client is closed`)) {
		t.Errorf("Get() error = %v, want = %v", err, nil)
	}

	d = *New(testInstance)
}

func TestDriverClose(t *testing.T) {
	d.Close()

	err := d.Set(valNotExist, valWithoutExpire, longExpire)
	if !reflect.DeepEqual(err, errors.New(`redis: client is closed`)) {
		t.Errorf("Get() error = %v, want = %v", err, nil)
	}

	d = *New(testInstance)
}

func TestNew(t *testing.T) {
	d := New(testInstance)

	err := d.storage.Ping().Err()
	if err != nil {
		t.Errorf("driver not created = %v", err)
	}
}
