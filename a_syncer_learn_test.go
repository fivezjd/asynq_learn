package asynq_learn

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// 重试任务

type reDo struct {
	fn       func() error
	deadline time.Time
	errMsg   error
}

type redoService struct {
	requests chan *reDo
}

func (r *redoService) Run() {
	request := make([]*reDo, 0)
	for {
		select {
		case req := <-r.requests:
			request = append(request, req)
		case <-time.After(time.Second * 5): // 5秒处理一次失败的任务
			temp := make([]*reDo, 0)
			for _, do := range request {
				err := do.fn()
				if err != nil {
					temp = append(temp, do)
				}
				fmt.Println(do.errMsg)
			}
			request = temp
		}
	}
}

func TestR(t *testing.T) {
	s := &redoService{
		requests: make(chan *reDo, 10),
	}
	go func() {
		for i := 0; i < 10; i++ {
			// mock 10个错误
			s.requests <- &reDo{
				fn: func() error {
					return nil
				},
				errMsg: errors.New("abc"),
			}
		}
	}()
	s.Run()

	select {}
}
