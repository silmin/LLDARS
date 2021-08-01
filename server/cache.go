package server

import (
	"log"
	"sync"
	"time"
)

type item struct {
	id      uint32
	expires int64
}

type IdCache struct {
	sync.Mutex
	items map[uint32]*item
}

func NewIdCache() *IdCache {
	c := &IdCache{items: make(map[uint32]*item)}
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				c.Lock()
				for k, v := range c.items {
					now := time.Now().UnixNano()
					if v.IsExpired(now) {
						log.Printf("%v has expires at %d\n", v, now)
						delete(c.items, k)
					}
				}
				c.Unlock()
			}
		}
	}()
	return c
}

func (i *item) IsExpired(t int64) bool {
	if i.expires == 0 {
		return true
	}
	return t > i.expires
}

func (c *IdCache) Put(id uint32, expires int64) {
	c.Lock()
	if _, ok := c.items[id]; !ok {
		c.items[id] = &item{
			id:      id,
			expires: expires,
		}
	}
	c.Unlock()
}

func (c *IdCache) Exists(id uint32) bool {
	c.Lock()
	_, ok := c.items[id]
	c.Unlock()

	if ok {
		return true
	} else {
		return false
	}
}
