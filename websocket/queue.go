package websocket

import (
	"sync"
)

// QueueNode ...
type QueueNode struct {
	data interface{}
	next *QueueNode
}

// Queue ...
type Queue struct {
	sync.Mutex

	head *QueueNode
	tail *QueueNode
	size int
}

// Len ...
func (q *Queue) Len() int {
	q.Lock()
	defer q.Unlock()
	return q.size
}

// Push ...
func (q *Queue) Push(item interface{}) {
	q.Lock()
	defer q.Unlock()

	n := &QueueNode{data: item}

	if q.tail == nil {
		q.tail = n
		q.head = n
	} else {
		q.tail.next = n
		q.tail = n
	}

	q.size++
}

// Pop ...
func (q *Queue) Pop() interface{} {
	q.Lock()
	defer q.Unlock()

	if q.head == nil {
		return nil
	}

	n := q.head
	q.head = n.next

	if q.head == nil {
		q.tail = nil
	}

	q.size--

	return n.data
}

// Peek ...
func (q *Queue) Peek() interface{} {
	q.Lock()
	defer q.Unlock()

	n := q.head
	if n == nil {
		return nil
	}

	return n.data
}
