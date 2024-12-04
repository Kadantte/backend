package context

import (
	"context"
	"sync"
)

type Context = context.Context
type CancelFunc = context.CancelFunc

func Background() Context {
	return context.Background()
}

func WithCancel(parent Context) (Context, CancelFunc) {
	return context.WithCancel(parent)
}

type waitGroupKeyType int
type waitGroupValueType = *sync.WaitGroup

const waitGroupKey waitGroupKeyType = 0

func WithWaitGroup(parent Context) Context {
	return context.WithValue(parent, waitGroupKey, new(sync.WaitGroup))
}

func GetWaitGroup(ctx Context) waitGroupValueType {
	return ctx.Value(waitGroupKey).(waitGroupValueType)
}
