package datastore

import (
	"sync"
	"time"

	"github.com/efficientgo/core/errors"
)

type SetCondition int

const (
	Default SetCondition = iota
	SetIfNotExists
	SetIfExists
)

type Element struct {
	value  interface{}
	expiry time.Duration
	ts     time.Time
}

type store struct {
	data map[interface{}]Element
	mut  sync.Mutex
}

type queue struct {
	list map[interface{}][]interface{}
	mut  sync.Mutex
}

type datastore struct {
	queue
	store
}

type DataStore interface {
	Set(key, value interface{}, expiry time.Duration, condition SetCondition) error
	Get(key interface{}) (interface{}, bool)
	QPush(key interface{}, value ...interface{}) error
	QPop(key interface{}) (interface{}, error)
	BQPop(key interface{}, duration time.Duration) (interface{}, error)
}

func NewDataStore() *datastore {
	return &datastore{
		store: store{data: make(map[interface{}]Element)},
		queue: queue{list: make(map[interface{}][]interface{})},
	}
}

func (d *datastore) Set(key, value interface{}, expiry time.Duration, condition SetCondition) error {
	d.store.mut.Lock()
	defer d.store.mut.Unlock()
	if value == nil {
		return errors.New("value can't be nil")
	}
	if d.data == nil {
		d.data = make(map[interface{}]Element)
	}
	elem, ok := d.data[key]

	switch condition {
	case SetIfNotExists:
		if !ok || time.Now().After(elem.ts.Add(elem.expiry*time.Second)) {
			d.data[key] = Element{value: value, ts: time.Now(), expiry: expiry}
			return nil
		}
	case SetIfExists:
		if ok && time.Now().Before(elem.ts.Add(elem.expiry*time.Second)) {
			d.data[key] = Element{value: value, ts: time.Now(), expiry: expiry}
			return nil
		}
	case Default:
		d.data[key] = Element{value: value, ts: time.Now(), expiry: expiry}
		return nil
	default:
		return errors.Newf("unsupported condition for set: %d", condition)
	}
	return nil
}

func (d *datastore) Get(key interface{}) (interface{}, bool) {
	elem, ok := d.data[key]
	if ok && time.Now().After(elem.ts.Add(elem.expiry*time.Second)) {
		return elem.value, false
	}
	return elem.value, ok
}

func (d *datastore) QPush(key interface{}, value ...interface{}) error {
	d.queue.mut.Lock()
	defer d.queue.mut.Unlock()
	if key == nil {
		return errors.New("invalid key - can't be nil")
	} else if len(value) == 0 {
		return errors.Newf("no value provided for the key \"%v\"", key)
	}
	if d.list == nil {
		d.list = make(map[interface{}][]interface{})
	}
	elem, ok := d.list[key]
	if !ok || elem == nil {
		d.list[key] = make([]interface{}, 0, 6)
	}
	d.list[key] = append(d.list[key], value...)
	return nil
}

func (d *datastore) QPop(key interface{}) (interface{}, error) {
	d.queue.mut.Lock()
	defer d.queue.mut.Unlock()
	if key == nil {
		return nil, errors.New("invalid key - can't be nil")
	}

	elem, ok := d.list[key]
	if !ok || elem == nil {
		return nil, nil
	}
	if len(elem) == 0 {
		return nil, nil
	}
	value := elem[len(elem)-1]
	d.list[key] = elem[:len(elem)-1]
	return value, nil
}

func (d *datastore) BQPop(key interface{}, duration time.Duration) (interface{}, error) {
	return nil, nil
}
