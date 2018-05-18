package counters

import (
	"sync"
	"sync/atomic"
)

// 一组带名字的计数器, 支持多线程

var mutex sync.Mutex
var counters = make(map[string]*int64)

func get(name string) (p *int64) {
	mutex.Lock()
	var ok bool
	if p, ok = counters[name]; !ok {
		p = new(int64)
		counters[name] = p
	}
	mutex.Unlock()
	return
}

func Set(name string, x int64) {
	atomic.StoreInt64(get(name), x)
}

func Add(name string, x int64) int64 {
	return atomic.AddInt64(get(name), x)
}

func Inc(name string) int64 {
	return atomic.AddInt64(get(name), 1)
}

func Dec(name string) int64 {
	return atomic.AddInt64(get(name), -1)
}

func Get(name string) int64 {
	return atomic.LoadInt64(get(name))
}

func Remove(name string) {
	mutex.Lock()
	delete(counters, name)
	mutex.Unlock()
}

func All() (m map[string]int64) {
	m = make(map[string]int64)
	mutex.Lock()
	for k, v := range counters {
		m[k] = *v
	}
	mutex.Unlock()
	return
}
