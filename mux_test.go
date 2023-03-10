package asynq

import (
	"io"
	"sync"
)

type handleF func(r io.Reader) error

type ServerMux struct {
	mu      sync.RWMutex       // 控制mu方法的原子性
	handMap map[string]handleF // 保存各个handle方法
}
