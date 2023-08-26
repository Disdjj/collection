package queue

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type queueNode struct {
	value any
	next  unsafe.Pointer
}

var nodePool = sync.Pool{
	New: func() any {
		return &queueNode{
			value: nil,
			next:  nil,
		}
	},
}

func newNode(v any) *queueNode {
	node := nodePool.Get().(*queueNode)
	node.value = v
	return node
}

type LockFreeQueue struct {
	cap  int32
	head unsafe.Pointer
	tail unsafe.Pointer
}

func (q *LockFreeQueue) Enqueue(v any) {
	node := newNode(v)
	for {
		tail := load(&q.tail)
		next := tail.next
		if tail == load(&q.tail) {
			if next == nil {
				// add to tail
				if atomic.CompareAndSwapPointer(&tail.next, next, unsafe.Pointer(node)) {
					// switch tail to node
					atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(node))
					atomic.AddInt32(&q.cap, 1)
					break
				}
			} else {
				// get wrong tail node
				// tail = tail.next
				// if can not , reload tail in next loop
				atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), next)
			}
		}
	}
}

func (q *LockFreeQueue) Dequeue() (res any) {
	for {
		head := load(&q.head)
		tail := load(&q.tail)
		next := head.next

		if head == load(&q.head) {
			if head == tail {
				if next == nil {
					return nil // has no value
				}
				atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), next)
			} else {
				node := load(&next)
				v := node.value
				if atomic.CompareAndSwapPointer(&q.head, unsafe.Pointer(head), next) {
					atomic.AddInt32(&q.cap, -1)
					defer nodePool.Put(node)
					return v
				}
			}
		}
	}
}

func (q *LockFreeQueue) Empty() bool {
	return atomic.LoadPointer(&q.head) == atomic.LoadPointer(&q.tail)
}

func (q *LockFreeQueue) Len() int32 {
	return atomic.LoadInt32(&q.cap)
}
func NewLockFreeQueue() *LockFreeQueue {
	node := unsafe.Pointer(&queueNode{})
	return &LockFreeQueue{
		head: node,
		tail: node,
	}
}

func load(p *unsafe.Pointer) *queueNode {
	return (*queueNode)(atomic.LoadPointer(p))
}
