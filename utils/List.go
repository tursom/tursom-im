package utils

import (
	"fmt"
	"strings"
)

type Iterator interface {
	HasNext() bool
	Next() interface{}
}

type MultipleIterator interface {
	HasNext() bool
	Next() interface{}
	Remove()
}

type List interface {
	Get(index uint) interface{}
	IndexOf(element interface{}) int
	Iterator() Iterator
}

func Contains(l List, element interface{}) bool {
	return l.IndexOf(element) >= 0
}

type MultipleList interface {
	Get(index uint) interface{}
	IndexOf(element interface{}) int
	Iterator() Iterator
	MultipleIterator() MultipleIterator
	Append(element interface{})
	Remove(element interface{}) interface{}
}

type LinkedList struct {
	first *linkedListNode
	tail  *linkedListNode
	size  uint
}

func (l LinkedList) String() string {
	if l.first == nil {
		return "[]"
	}

	builder := strings.Builder{}
	builder.WriteString("[")
	iterator := l.Iterator()
	next := iterator.Next()
	builder.WriteString(fmt.Sprint(next))
	for iterator.HasNext() {
		builder.WriteString(", ")
		next = iterator.Next()
		builder.WriteString(fmt.Sprint(next))
	}
	builder.WriteString("]")
	return builder.String()
}

type linkedListNode struct {
	value interface{}
	next  *linkedListNode
}

type linkedListIterator struct {
	node *linkedListNode
}

func (iter *linkedListIterator) HasNext() bool {
	return iter.node != nil
}

func (iter *linkedListIterator) Next() interface{} {
	value := iter.node.value
	iter.node = iter.node.next
	return value
}

func (l LinkedList) Iterator() Iterator {
	return &linkedListIterator{l.first}
}

type linkedListMultipleIterator struct {
	list  *LinkedList
	pprev *linkedListNode
	prev  *linkedListNode
	node  *linkedListNode
}

func (iter *linkedListMultipleIterator) HasNext() bool {
	return iter.node != nil
}

func (iter *linkedListMultipleIterator) Next() interface{} {
	iter.pprev = iter.prev
	iter.prev = iter.node
	iter.node = iter.node.next
	return iter.prev.value
}

func (iter *linkedListMultipleIterator) Remove() {
	if iter.prev == nil {
		return
	}
	if iter.pprev == nil {
		iter.list.first = iter.node
		if iter.node == nil {
			iter.list.tail = nil
		}
	} else {
		iter.pprev.next = iter.node
	}
}

func (l *LinkedList) MultipleIterator() MultipleIterator {
	return &linkedListMultipleIterator{l, nil, nil, l.first}
}

func NewLinkedList() *LinkedList {
	return &LinkedList{
		first: nil,
		tail:  nil,
		size:  0,
	}
}

func (l LinkedList) Get(index uint) interface{} {
	node := l.first
	for node != nil {
		if index == 0 {
			return node.value
		}
		node = node.next
		index--
	}
	return nil
}

func (l LinkedList) IndexOf(element interface{}) int {
	var index int = 0
	node := l.first
	for node != nil {
		if node.value == element {
			return index
		}
		node = node.next
		index++
	}
	return -1
}

func (l *LinkedList) Append(element interface{}) {
	if l.tail == nil {
		l.first = &linkedListNode{
			value: element,
		}
		l.tail = l.first
	} else {
		l.tail.next = &linkedListNode{
			value: element,
		}
		l.tail = l.tail.next
	}
}

func (l *LinkedList) Remove(element interface{}) interface{} {
	var index uint = 0
	node := l.first
	var prev *linkedListNode = nil
	for node != nil {
		if node.value == element {
			if prev != nil {
				prev.next = node.next
			} else {
				l.first = node.next
			}
			return node.value
		}
		prev = node
		node = node.next
		index++
	}
	return -1
}
