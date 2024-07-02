package fs

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/kxnes/go-interviews/apicache/test"
)

const (
	keyExist = "exist"
	valExist = "exist"
	ttlExist = 10
	maxConn  = 10
	timeout  = 10
)

var f = fileSystem{
	driver: &test.DriverMock{Storage: &sync.Map{}},
	opts: &Options{
		MaxConn: maxConn,
		Timeout: timeout,
	},
	done:  make(chan struct{}),
	queue: make(chan struct{}, maxConn),
}

func TestMain(m *testing.M) {
	_ = f.driver.Set(keyExist, valExist, ttlExist)
	os.Exit(m.Run())
}

func TestFileSystemGet(t *testing.T) {
	cases := []struct {
		name string
		key  string
		want string
		err  error
	}{
		{
			name: "empty key",
			key:  "",
			want: "",
			err:  &ErrEmptyKey{},
		},
		{
			name: "key not exist",
			key:  test.KeyNotExist,
			want: "",
			err:  &ErrNotExist{"sd"},
		},
		{
			name: "internal error",
			key:  test.KeyError,
			want: "",
			err:  fmt.Errorf(ErrKVStorage, errors.New(test.InternalError)),
		},
		{
			name: "key exist",
			key:  keyExist,
			want: valExist,
			err:  nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := f.Get(c.key)

			if got != c.want {
				t.Errorf("Get() got = %v, want = %v", got, c.want)
			}

			if !errors.As(err, &c.err) {
				if (c.err == nil) == (err == nil) {
					return
				}

				t.Errorf("Get() error = %v, want = %v", err, c.err)
			}
		})
	}
}

func TestFileSystemSet(t *testing.T) {
	cases := []struct {
		name string
		key  string
		val  string
		ttl  int
		err  error
	}{
		{
			name: "empty key",
			key:  "",
			val:  "",
			ttl:  0,
			err:  &ErrEmptyKey{},
		},
		{
			name: "empty val",
			key:  keyExist,
			val:  "",
			ttl:  0,
			err:  &ErrEmptyVal{test.KeyNotExist},
		},
		{
			name: "negative ttl",
			key:  keyExist,
			val:  valExist,
			ttl:  -10,
			err:  &ErrInvalidTTL{keyExist, -10},
		},
		{
			name: "zero ttl",
			key:  keyExist,
			val:  valExist,
			ttl:  0,
			err:  &ErrInvalidTTL{keyExist, 0},
		},
		{
			name: "internal error",
			key:  test.KeyError,
			val:  valExist,
			ttl:  ttlExist,
			err:  fmt.Errorf(ErrKVStorage, errors.New(test.InternalError)),
		},
		{
			name: "normal set",
			key:  keyExist,
			val:  valExist,
			ttl:  ttlExist,
			err:  nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := f.Set(c.key, c.val, c.ttl)

			if !errors.As(err, &c.err) {
				if (c.err == nil) == (err == nil) {
					return
				}

				t.Errorf("Set() error = %v, want = %v", err, c.err)
			}
		})
	}
}

func TestFileSystemDelete(t *testing.T) {
	cases := []struct {
		name string
		key  string
		want bool
		err  error
	}{
		{
			name: "empty key",
			key:  "",
			want: false,
			err:  &ErrEmptyKey{},
		},
		{
			name: "key not exist",
			key:  test.KeyNotExist,
			want: false,
			err:  &ErrNotExist{test.KeyNotExist},
		},
		{
			name: "internal error",
			key:  test.KeyError,
			want: false,
			err:  fmt.Errorf(ErrKVStorage, errors.New(test.InternalError)),
		},
		{
			name: "normal delete",
			key:  keyExist,
			want: true,
			err:  nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := f.Delete(c.key)

			if got != c.want {
				t.Errorf("Delete() got = %v, want = %v", got, c.want)
			}

			if !errors.As(err, &c.err) {
				if (c.err == nil) == (err == nil) {
					return
				}

				t.Errorf("Delete() error = %v, want = %v", err, c.err)
			}
		})
	}
}

func TestFileSystemClose(t *testing.T) {
	f.Close()

	_, ok := <-f.done
	if ok {
		t.Errorf("Close() done channel not closed")
	}

	if f.driver.(*test.DriverMock).Closed != 1 {
		t.Errorf("Close() storage not closed")
	}

	f.driver.(*test.DriverMock).Closed = 0
	f.done = make(chan struct{})
}

func TestNew(t *testing.T) {
	got := New(
		&test.DriverMock{Storage: &sync.Map{}},
		&Options{
			MaxConn: maxConn,
			Timeout: timeout,
		},
	).(*fileSystem)

	want := &fileSystem{
		driver: &test.DriverMock{Storage: &sync.Map{}},
		opts: &Options{
			MaxConn: maxConn,
			Timeout: timeout,
		},
	}

	if !reflect.DeepEqual(got.driver, want.driver) {
		t.Errorf("New() = %v, want = %v", got, *want)
	}

	if cap(got.queue) != maxConn {
		t.Errorf("New() queue = %d, want = %d", cap(got.queue), maxConn)
	}

	if cap(got.done) != 0 {
		t.Errorf("New() done chan = %d, want = %d", cap(got.done), 0)
	}

	if !reflect.DeepEqual(got.opts, want.opts) {
		t.Errorf("New() = %v, want = %v", got, *want)
	}
}

func TestFileSystemErrors(t *testing.T) {
	ect := &ErrConcurrentTimeout{op: "get"}
	ectW := "timeout for operation (get)"

	if ect.Error() != ectW {
		t.Errorf("ErrConcurrentTimeout.Error() = %s, want = %s", ect.Error(), ectW)
	}

	ecd := &ErrCloseDriver{}
	ecdW := "driver is closed"

	if ecd.Error() != ecdW {
		t.Errorf("ErrCloseDriver.Error() = %s, want = %s", ecd.Error(), ecdW)
	}

	eek := &ErrEmptyKey{}
	eekW := "empty key"

	if eek.Error() != eekW {
		t.Errorf("ErrEmptyKey.Error() = %s, want = %s", eek.Error(), eekW)
	}

	eev := &ErrEmptyVal{key: test.KeyError}
	eevW := "empty value for key (" + test.KeyError + ")"

	if eev.Error() != eevW {
		t.Errorf("ErrEmptyVal.Error() = %s, want = %s", eev.Error(), eevW)
	}

	eit := &ErrInvalidTTL{key: test.KeyNotExist, ttl: -10}
	eitW := fmt.Sprintf("invalid ttl (%d) for key (%s)", -10, test.KeyNotExist)

	if eit.Error() != eitW {
		t.Errorf("ErrInvalidTTL.Error() = %s, want = %s", eit.Error(), eitW)
	}

	ene := &ErrNotExist{key: test.KeyNotExist}
	eneW := "key (" + test.KeyNotExist + ") not exist"

	if ene.Error() != eneW {
		t.Errorf("ErrNotExist.Error() = %s, want = %s", ene.Error(), eneW)
	}
}

func TestFileSystemConcurrent(t *testing.T) {
	// Check acquire and release on operations

	fs := &fileSystem{
		driver: &test.DriverMock{
			Storage:      &sync.Map{},
			IsConcurrent: true,
		},
		opts: &Options{
			Timeout: 10,
		},
		done:  make(chan struct{}),
		queue: make(chan struct{}, 3),
	}

	go func() { _, _ = fs.Get(test.KeyError) }()
	go func() { _ = fs.Set(test.KeyError, valExist, ttlExist) }()
	go func() { _, _ = fs.Delete(test.KeyError) }()

	time.Sleep(time.Second)

	if len(fs.queue) != 3 {
		t.Errorf("acquare not happened")
	}

	time.Sleep(2 * time.Second)
	if len(fs.queue) != 0 {
		t.Errorf("release not happened")
	}

	// Check dummy connections will be done after close

	for i := 0; i < 10; i++ {
		go func(i int) {
			if i%3 == 0 {
				time.Sleep(4 * time.Second)
			}
			_ = fs.Set(strconv.Itoa(i), valExist, ttlExist)
		}(i)
	}

	time.Sleep(2 * time.Second)

	fs.Close()

	storage := fs.driver.(*test.DriverMock).Storage

	for i := 0; i < 10; i++ {
		_, ok := storage.Load(strconv.Itoa(i))

		if ok {
			if i%3 == 0 {
				t.Errorf("key (%d) set but doesn't", i)
			}
			continue
		}

		if !ok {
			if i%3 != 0 {
				t.Errorf("key (%d) doesn't set but must", i)
			}
		}
	}

	if len(fs.queue) != 0 {
		fmt.Println("queue not empty")
	}
}

func TestFileSystemConcurrentErrors(t *testing.T) {
	// Timeout if queue is full

	fs := &fileSystem{
		driver: &test.DriverMock{
			Storage:      &sync.Map{},
			IsConcurrent: true,
		},
		opts: &Options{
			Timeout: 1,
		},
		done:  make(chan struct{}),
		queue: make(chan struct{}, 2),
	}

	fs.queue <- struct{}{}
	fs.queue <- struct{}{}

	var ect *ErrConcurrentTimeout

	got, err := fs.Get(keyExist)
	if got != "" || !errors.As(err, &ect) {
		t.Errorf("Get() timeout not happened")
	}

	err = fs.Set(keyExist, valExist, ttlExist)
	if got != "" || !errors.As(err, &ect) {
		t.Errorf("Set() timeout not happened")
	}

	ok, err := fs.Delete(keyExist)
	if ok || !errors.As(err, &ect) {
		t.Errorf("Delete() timeout not happened")
	}

	<-fs.queue
	<-fs.queue

	// Error after closed

	fs.Close()

	got, err = fs.Get(keyExist)

	var ecd *ErrCloseDriver
	if got != "" || !errors.As(err, &ecd) {
		t.Errorf("oparation happened")
	}
}
