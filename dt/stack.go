package dt

import (
	"sync"
	"sync/atomic"
)

// 分类：顺序栈和链式栈
// 定义：操作受限的线性表，只允许一端插入和删除
// 特性:
// 场景:

type StackLock struct {
	data   []interface{}
	length int32 //栈大小
	count  int32 //栈元素
	lock   *sync.Mutex
}

func NewStack(length int32) *StackLock {
	return &StackLock{
		data:   make([]interface{}, length),
		length: length,
		count:  0,
		lock:   &sync.Mutex{},
	}
}

func (s *StackLock) Pop() interface{} {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.count <= 0 {
		return nil
	}
	s.count--
	return s.data[s.count]
}

func (s *StackLock) Push(val interface{}) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.count >= s.length {
		return false
	}
	s.data[s.count] = val
	s.count++
	return true
}

type StackCAS struct {
	data   []interface{}
	length int32 //栈大小
	count  int32 //栈元素
}

func NewStackCAS(length int32) *StackCAS {
	return &StackCAS{
		data:   make([]interface{}, length),
		length: length,
		count:  0,
	}
}

func (s *StackCAS) Pop() interface{} {
	var old int32
	for {
		old = s.count
		if old <= 0 {
			return nil
		}
		if atomic.CompareAndSwapInt32(&s.count, old, old-1) {
			break
		}
	}
	return s.data[old-1]
}

func (s *StackCAS) Push(val interface{}) bool {
	var old int32
	for {
		old = s.count
		if s.length <= old {
			return false
		}
		if atomic.CompareAndSwapInt32(&s.count, old, old+1) {
			break
		}
	}

	s.data[old] = val
	return true
}
