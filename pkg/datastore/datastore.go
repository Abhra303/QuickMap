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
	Set(key, value interface{}, expiry time.Duration, condition SetCondition) (bool, error)
	Get(key interface{}) interface{}
	QPush(key interface{}, value []interface{}) error
	QPop(key interface{}) (interface{}, error)
	BQPop(key interface{}, duration time.Duration) (interface{}, error)
}

func NewDataStore() *datastore {
	return &datastore{
		store: store{data: make(map[interface{}]Element)},
		queue: queue{list: make(map[interface{}][]interface{})},
	}
}

func (d *datastore) Set(key, value interface{}, expiry time.Duration, condition SetCondition) (bool, error) {
	d.store.mut.Lock()
	defer d.store.mut.Unlock()
	if value == nil {
		return false, errors.New("value can't be nil")
	}
	if d.data == nil {
		d.data = make(map[interface{}]Element)
	}
	elem, ok := d.data[key]

	switch condition {
	case SetIfNotExists:
		if !ok || (elem.expiry != 0 && time.Now().After(elem.ts.Add(elem.expiry*time.Second))) {
			d.data[key] = Element{value: value, ts: time.Now(), expiry: expiry}
			return true, nil
		}
	case SetIfExists:
		if ok && (elem.expiry == 0 || time.Now().Before(elem.ts.Add(elem.expiry*time.Second))) {
			d.data[key] = Element{value: value, ts: time.Now(), expiry: expiry}
			return true, nil
		}
	case Default:
		d.data[key] = Element{value: value, ts: time.Now(), expiry: expiry}
		return true, nil
	default:
		return false, errors.Newf("unsupported condition for set: %d", condition)
	}
	return false, nil
}

func (d *datastore) Get(key interface{}) interface{} {
	elem, ok := d.data[key]
	if ok && (elem.expiry == 0 || time.Now().Before(elem.ts.Add(elem.expiry*time.Second))) {
		return elem.value
	}
	return nil
}

func (d *datastore) QPush(key interface{}, values []interface{}) error {
	if key == nil {
		return errors.New("invalid key - can't be nil")
	} else if len(values) == 0 {
		return errors.Newf("no value provided for the key \"%v\"", key)
	}
	d.queue.mut.Lock()
	defer d.queue.mut.Unlock()
	if d.list == nil {
		d.list = make(map[interface{}][]interface{})
	}
	elem, ok := d.list[key]
	if !ok || elem == nil {
		d.list[key] = make([]interface{}, 0, 6)
	}
	d.list[key] = append(d.list[key], values...)
	return nil
}

func (d *datastore) QPop(key interface{}) (interface{}, error) {
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
	d.queue.mut.Lock()
	defer d.queue.mut.Unlock()
	value := elem[len(elem)-1]
	d.list[key] = elem[:len(elem)-1]
	return value, nil
}

func (d *datastore) BQPop(key interface{}, duration time.Duration) (interface{}, error) {
	if key == nil {
		return nil, errors.New("invalid key - can't be nil")
	}

	d.queue.mut.Lock()
	defer d.queue.mut.Unlock()
	if len(d.list[key]) == 0 {
		if duration == 0 {
			return nil, nil
		}

		timer := time.NewTimer(duration * time.Second)

		for {
			// not the most efficient way to do this.
			// sync.Cond could be a better solution. But seems like
			// it is problemetic to use with timer check.
			d.queue.mut.Unlock()
			time.Sleep(1 * time.Second)
			d.queue.mut.Lock()
			select {
			case <-timer.C:
				return nil, nil
			default:
			}
			if len(d.list[key]) > 0 {
				break
			}

		}
		if !timer.Stop() {
			<-timer.C
		}
	}

	value := d.list[key][len(d.list[key])-1]
	d.list[key] = d.list[key][:len(d.list[key])-1]
	return value, nil
}
