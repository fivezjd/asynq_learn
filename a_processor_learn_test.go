package asynq_learn

import (
	"fmt"
	"testing"
)

// 一个接口类型
// 函数类型
//  这种设计方式可以方便地将已有的函数类型作为接口类型的实现使用，而不需要为每个函数类型都定义一个新的结构体类型来实现接口
//

type A interface {
	Do() error
}

func C(a A) error {
	return a.Do()
}

// 如果不使用hand强制类型转换，每个传入的类型必须得手动实现Do 方法，很不灵活
type hand func() error

func (receiver hand) Do() error {
	return receiver()
}

func TestName3(t *testing.T) {
	C(hand(func() error {
		fmt.Println("do")
		return nil
	}))
}
