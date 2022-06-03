package ElasticCache

import (
	"sync"
	"time"
)

type ElasticCache interface {

	// GetAndSet key缓存数据的key limitTime 数据存在的有效期，get获取数据的处理函数（如果过期了会删除该key的数据释放内存）。
	GetAndSet(key string, limitTime time.Duration, get getter) (data interface{})
	// Delete 删除数据
	Delete(key string)
	// Clear 清理缓存
	Clear()
}

func New(clearLimit time.Duration) ElasticCache {
	e := &elasticCache{
		mu:         &sync.Mutex{},
		clearlimit: clearLimit,
		caches:     make(map[string]*cache),
	}
	go e.run()
	return e
}

type elasticCache struct {
	mu         *sync.Mutex
	clearlimit time.Duration
	caches     map[string]*cache
	clearCh    <-chan struct{}
	isClose    bool
}

type getter func(key string) (data interface{}, whetherCache bool)

func (e *elasticCache) GetAndSet(key string, limitTime time.Duration, getDataFn getter) (data interface{}) {
	if e.isClose {
		return nil
	}
	if getDataFn == nil {
		panic("getter should be not nil")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	now := time.Now()
	c, ok := e.caches[key]
	if !ok { //如果缓存不存在就从处理函数获取
		d, isCache := getDataFn(key)
		if isCache {
			w := &cache{
				lastTime:  now,
				limitTime: limitTime,
				key:       key,
				handle:    getDataFn,
				getNum:    1,
				Data:      d,
			}
			e.caches[key] = w
		}
		return d
	}
	if now.Sub(c.lastTime) > c.limitTime { //如果缓存数据已经过期就从处理函数获取并刷新数据
		var d interface{}
		d, setCache := c.handle(key)
		if setCache {
			c.lastTime = time.Now()
			c.Data = d
			c.getNum = 1
		} else {
			delete(e.caches, key)
		}
		return d
	}
	c.getNum++
	return c.Data
}

func (e *elasticCache) run() {
	ticker := time.NewTicker(e.clearlimit)
	for {
		select {
		case <-ticker.C:
			for k, v := range e.caches {
				now := time.Now()
				if now.Sub(v.lastTime) > v.limitTime {
					e.mu.Lock()
					delete(e.caches, k)
					e.mu.Unlock()
				}
			}
		case <-e.clearCh:
			e.mu.Lock()
			e.isClose = true
			for k := range e.caches {
				delete(e.caches, k)
			}
			e.mu.Unlock()
			return
		}
	}
}

func (e *elasticCache) Clear() {
	e.clearCh = make(chan struct{}, 1)
}

func (e *elasticCache) Set(key string, limitTime time.Duration, handle getter) {
	e.mu.Lock()
	defer e.mu.Unlock()
	data, setCache := handle(key)
	if setCache {
		e.caches[key] = &cache{
			key:       key,
			Data:      data,
			getNum:    0,
			limitTime: limitTime,
			lastTime:  time.Now(),
		}
	}

}

func (e *elasticCache) Delete(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.caches, key)
}

type cache struct {
	lastTime  time.Time
	limitTime time.Duration
	handle    getter
	getNum    int
	key       string
	Data      interface{}
}
