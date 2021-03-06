package dt

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// 分类：顺序栈和链式栈
// 定义：操作受限的线性表，只允许一端插入和删除
// 特性:
// 场景:

type Stack interface {
	Pop() interface{}
	Push(interface{}) bool
}
type STACK int

const (
	LockStack STACK = iota
	CasStack
	SliceStack
	LinkStack
)

type StackSlice struct {
	data []interface{}
}

func (s *StackSlice) Pop() interface{} {
	length := len(s.data)
	if length == 0 {
		return nil
	}
	x := s.data[length-1]
	s.data = s.data[:length-1]
	return x
}

func (s *StackSlice) Push(i interface{}) bool {
	s.data = append(s.data, i)
	return true
}

type StackLock struct {
	data   []interface{}
	length int32 //栈大小
	count  int32 //栈元素
	lock   *sync.Mutex
}

func NewStack(length int32, tp STACK) Stack {
	switch tp {
	case LockStack:
		return &StackLock{
			data:   make([]interface{}, length),
			length: length,
			count:  0,
			lock:   &sync.Mutex{},
		}
	case CasStack:
		return &StackCAS{
			data:   make([]interface{}, length),
			length: length,
			count:  0,
		}
	case SliceStack:
		return &StackSlice{data: make([]interface{}, 0, length)}
	case LinkStack:
		return &StackLink{
			head: unsafe.Pointer(&Node{nil, nil}),
		}

	}
	panic("not found stack")

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
	val := s.data[old-1]
	return val
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
	// 栈顶先变化，数据延后更新，线程可能会被切走，导致pop出nil数据
	s.data[old] = val
	return true
}

type Node struct {
	val  interface{}
	prev unsafe.Pointer
}

type StackLink struct {
	head unsafe.Pointer
}

func (s *StackLink) Pop() interface{} {
	for {
		old := s.head
		x := (*Node)(atomic.LoadPointer(&old))
		if x.val == nil || x.prev == nil {
			return nil
		}

		val := x.val
		if atomic.CompareAndSwapPointer(&s.head, old, x.prev) {
			return val
		}
	}
}

func (s *StackLink) Push(val interface{}) bool {
	node := &Node{val: val, prev: nil}
	for {
		old := s.head
		node.prev = s.head
		n := unsafe.Pointer(node)
		if atomic.CompareAndSwapPointer(&s.head, old, n) {
			return true
		}
	}
}
