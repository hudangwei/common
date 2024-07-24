package dsl

import (
	"container/list"
)

type Stack struct {
	list *list.List
}

func NewStack() *Stack {
	return &Stack{list: list.New()}
}

func (stack *Stack) pop() interface{} {
	// 出栈
	e := stack.list.Back()
	if e != nil {
		stack.list.Remove(e)
		return e.Value
	}
	return nil
}

func (stack *Stack) push(v interface{}) {
	// 入栈
	stack.list.PushBack(v)
}

func (stack *Stack) isEmpty() bool {
	return stack.list.Len() == 0
}

func (stack *Stack) top() interface{} {
	e := stack.list.Back()
	if e != nil {
		return e.Value
	}
	return nil
}
