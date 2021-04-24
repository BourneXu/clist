package simpleList

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type IntList struct {
	head   *intNode
	length int
}

type intNode struct {
	value  int
	next   *intNode
	marked bool
	mu     sync.RWMutex
}

func newIntNode(value int) *intNode {
	return &intNode{value: value}
}

func NewInt() *IntList {
	return &IntList{head: newIntNode(0)}
}

func (n *intNode) loadNext() *intNode {
	return (*intNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&n.next))))
}

func (n *intNode) storeNext(node *intNode) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&n.next)), unsafe.Pointer(node))
}

func (l *IntList) Insert(value int) bool {
	for {
		a := l.head
		b := a.next
		for b != nil && b.value < value {
			a = b
			b = b.loadNext()
		}
		// Check if the node is exist.
		if b != nil && b.value == value {
			return false
		}
		// lock A and check if A.next!= B or A.marked, if true, unlock A and continue.
		a.mu.Lock()
		if a.next != b || a.marked {
			a.mu.Unlock()
			continue
		}
		defer a.mu.Unlock()
		x := newIntNode(value)
		x.storeNext(b)
		a.storeNext(x)
		l.length++
		break
	}
	return true
}

func (l *IntList) Delete(value int) bool {
	for {
		a := l.head
		b := a.next
		for b != nil && b.value < value {
			a = b
			b = b.loadNext()
		}
		// Check if b is not exists
		if b == nil || b.value != value {
			return false
		}
		// Lock B and check if B.marked is true, then continue
		b.mu.Lock()
		if b.marked {
			b.mu.Unlock()
			continue
		}
		// Lock A and check if A.marked is true or A.next != B, then continue
		a.mu.Lock()
		if a.marked || a.next != b {
			a.mu.Unlock()
			b.mu.Unlock()
			continue
		}
		defer a.mu.Unlock()
		defer b.mu.Unlock()
		b.marked = true
		a.storeNext(b.loadNext())
		l.length--
		break
	}
	return true
}

func (l *IntList) Contains(value int) bool {
	x := l.head.next
	for x != nil && x.value < value {
		x = x.loadNext()
	}
	if x == nil {
		return false
	}
	return (x.value == value) && x.marked
}

func (l *IntList) Range(f func(value int) bool) {
	x := l.head.next
	for x != nil {
		if !f(x.value) {
			break
		}
		x = x.next
	}
}

func (l *IntList) Len() int {
	return int(l.length)
}
