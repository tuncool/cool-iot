package cache

import (
	"container/list"
	"sync"
)

type Cache struct {
	MaxEntries int
	OnEvicted  func(key Key, value interface{})
	ll         *list.List // list
	cache      sync.Map
}

type Key interface{}

type entry struct {
	key   Key
	value interface{}
}

func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
	}
}

func (c *Cache) Add(key Key, value interface{}) {
	if ee, ok := c.cache.Load(key); ok {
		c.ll.MoveToFront(ee.(*list.Element)) // move to the front
		ee.(*list.Element).Value.(*entry).value = value
		return
	}
	ele := c.ll.PushFront(&entry{key, value})
	c.cache.Store(key, ele)
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries { // Remove the oldest element if the limit is exceeded
		c.RemoveOldest()
	}
}

func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if ele, hit := c.cache.Load(key); hit {
		c.ll.MoveToFront(ele.(*list.Element))
		return ele.(*list.Element).Value.(*entry).value, true
	}
	return
}
func (c *Cache) GetString(key Key) (value string, ok bool) {
	v, ok := c.Get(key)
	if ok {
		value, ok = v.(string)
	}
	return
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key Key) {
	if ele, hit := c.cache.Load(key); hit {
		c.removeElement(ele.(*list.Element))
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	c.cache.Delete(kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	return c.ll.Len()
}

// Clear purges all stored items from the cache.
func (c *Cache) Clear() {
	if c.OnEvicted != nil {
		c.cache.Range(func(key, value interface{}) bool {
			kv := value.(*list.Element).Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
			return true
		})
	}
	c.ll = nil
}
