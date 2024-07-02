package memcache

import (
	"errors"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

const (
	testInstance = ":11211"

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

var d = Driver{storage: memcache.New(testInstance)}

func TestMain(m *testing.M) {
	for _, item := range []*memcache.Item{
		{
			Key:   keyWithoutExpire,
			Value: []byte(valWithoutExpire),
		},
		{
			Key:        keyWithLongExpire,
			Value:      []byte(valWithLongExpire),
			Expiration: longExpire,
		},
		{
			Key:        keyWithShortExpire,
			Value:      []byte(valWithShortExpire),
			Expiration: shortExpire,
		},
	} {
		err := d.storage.Set(item)
		if err != nil {
			log.Fatalf("setup fixture %v error = %v\n", *item, err)
		}
	}

	code := m.Run()

	err := d.storage.FlushAll()
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
	val, err := d.Get(valNotExist)

	if val != valNotExist {
		t.Errorf("Get() = %s, want = %s", val, valNotExist)
	}

	if !reflect.DeepEqual(err, errors.New(`memcache: unexpected line in get response: "ERROR\r\n"`)) {
		t.Errorf("Get() error = %v, want = %v", err, nil)
	}
}

func TestDriverSetNotExistKey(t *testing.T) {
	item, err := d.storage.Get(keyNotExist)
	if !errors.Is(err, memcache.ErrCacheMiss) || item != nil {
		t.Errorf("Key exists but doesn't")
	}

	err = d.Set(keyNotExist, valWithoutExpire, withoutExpire)

	if err != nil {
		t.Errorf("Set() error = %v, want = %v", err, nil)
	}

	item, err = d.storage.Get(keyNotExist)
	if errors.Is(err, memcache.ErrCacheMiss) || item == nil {
		t.Errorf("Set() doesn't set")
	}

	if string(item.Value) != valWithoutExpire {
		t.Errorf("Set() wrong value = %s", item.Value)
	}
}

func TestDriverSetExistKey(t *testing.T) {
	item, err := d.storage.Get(keyWithoutExpire)
	if errors.Is(err, memcache.ErrCacheMiss) || item == nil {
		t.Errorf("Key not exists but doesn't")
	}

	err = d.Set(keyWithoutExpire, valWithoutExpire, withoutExpire)

	if err != nil {
		t.Errorf("Set() error = %v, want = %v", err, nil)
	}

	item, err = d.storage.Get(keyWithoutExpire)
	if errors.Is(err, memcache.ErrCacheMiss) || item == nil {
		t.Errorf("Set() doesn't set")
	}

	if string(item.Value) != valWithoutExpire {
		t.Errorf("Set() wrong value = %s", item.Value)
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

	item, err := d.storage.Get(keyWithLongExpire)
	if errors.Is(err, memcache.ErrCacheMiss) || item == nil {
		t.Errorf("Set() doesn't set")
	}

	if string(item.Value) != valWithLongExpire {
		t.Errorf("Set() wrong value = %s", item.Value)
	}

	time.Sleep(time.Second)

	item, err = d.storage.Get(keyWithLongExpire)
	if !errors.Is(err, memcache.ErrCacheMiss) || item != nil {
		t.Errorf("Set() doesn't update expiration")
	}
}

func TestDriverSetNegativeTTL(t *testing.T) {
	err := d.Set(keyWithoutExpire, valWithoutExpire, invalidExpire)
	if err != nil {
		t.Errorf("Set() error = %v, want = %v", err, nil)
	}

	item, err := d.storage.Get(keyWithoutExpire)
	if !errors.Is(err, memcache.ErrCacheMiss) || item != nil {
		t.Errorf("Set() set but doesn't")
	}
}

func TestDriverSetUnexpectedError(t *testing.T) {
	err := d.Set(valNotExist, valWithoutExpire, invalidExpire)

	if !reflect.DeepEqual(err, errors.New(`memcache: unexpected response line from "set": "ERROR\r\n"`)) {
		t.Errorf("Get() error = %v, want = %v", err, nil)
	}
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
	ok, err := d.Delete(valNotExist)
	if ok {
		t.Errorf("Delete() = %v, want = %v", ok, false)
	}

	if !reflect.DeepEqual(err, errors.New(`memcache: unexpected response line: "ERROR\r\n"`)) {
		t.Errorf("Delete() error = %v, want = %v", err, nil)
	}
}

func TestDriverClose(t *testing.T) {
	d.Close()
}

func TestNew(t *testing.T) {
	d := New(testInstance)

	err := d.storage.Ping()
	if err != nil {
		t.Errorf("driver not created = %v", err)
	}
}
