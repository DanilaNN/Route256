package cache

import (
	"container/list"
	"sync"
)

type Item[KeyType comparable] struct {
	Key   KeyType
	Value interface{}
}

type LRU[KeyType comparable] struct {
	capacity int
	queue    *list.List
	mutex    *sync.RWMutex
	items    map[KeyType]*list.Element
}

func NewLRU[KeyType comparable](capacity int) *LRU[KeyType] {
	return &LRU[KeyType]{
		capacity: capacity,
		queue:    list.New(),
		mutex:    new(sync.RWMutex),
		items:    make(map[KeyType]*list.Element),
	}
}

func (c *LRU[KeyType]) Add(key KeyType, value interface{}) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if element, exists := c.items[key]; exists {
		c.queue.MoveToFront(element)
		element.Value.(*Item[KeyType]).Value = value
		return true
	}

	if c.queue.Len() == c.capacity {
		c.clear()
	}

	item := &Item[KeyType]{
		Key:   key,
		Value: value,
	}

	element := c.queue.PushFront(item)
	c.items[item.Key] = element

	return true
}

func (c *LRU[KeyType]) Get(key KeyType) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	element, exists := c.items[key]
	if !exists {
		return nil
	}

	c.queue.MoveToFront(element)
	return element.Value.(*Item[KeyType]).Value
}

func (c *LRU[KeyType]) Remove(key KeyType) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if val, found := c.items[key]; found {
		c.deleteItem(val)
	}

	return true
}

func (c *LRU[KeyType]) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

func (c *LRU[KeyType]) clear() {
	if element := c.queue.Back(); element != nil {
		c.deleteItem(element)
	}
}

func (c *LRU[KeyType]) deleteItem(element *list.Element) {
	item := c.queue.Remove(element).(*Item[KeyType])
	delete(c.items, item.Key)
}
