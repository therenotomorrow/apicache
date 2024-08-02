package machine

import (
	"context"
	"sync"

	"github.com/therenotomorrow/apicache/pkg/drivers"
)

type Machine struct {
	data  map[string]string
	mutex sync.RWMutex
}

func New() *Machine {
	return &Machine{data: make(map[string]string), mutex: sync.RWMutex{}}
}

func (d *Machine) Get(_ context.Context, key string) (string, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	val, ok := d.data[key]

	if !ok {
		return "", drivers.ErrNotExist
	}

	return val, nil
}

func (d *Machine) Set(_ context.Context, key string, val string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.data[key] = val

	return nil
}

func (d *Machine) Del(_ context.Context, key string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.data, key)

	return nil
}

func (d *Machine) Close() error {
	return nil
}
