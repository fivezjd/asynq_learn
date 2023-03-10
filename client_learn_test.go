package asynq

import (
	"context"
	"testing"
)

// go test -v -run client_learn_test.go
func TestNewClient(t *testing.T) {
	client := NewClient(RedisClientOpt{Addr: "localhost:6379"})
	t1 := NewTask("测试传入异类配置Option 入队", nil, MaxRetry(10))
	taskInfo, err := client.EnqueueContext(context.Background(), t1, ErrOption{}, MaxRetry(10))
	if err != nil {
		t.Error(err)
	}
	t.Log(client, taskInfo)
}

// 这个option不一定是个结构体，可以是其他自定义类型:
// type a int
// type b string
// type c map[string]string
type ErrOption struct {
	//Option
}

// 验证ErrOption 是否实现了Option 接口
var _ Option = (*ErrOption)(nil)

func (e ErrOption) String() string {
	return "ErrOption"
}

func (e ErrOption) Value() interface{} {
	return nil
}

func (e ErrOption) Type() OptionType {
	return 1
}

type c map[string]string

// 测试任务ID冲突

func TestConflictTask(t *testing.T) {
	// 任务配置可以在入队的时候指定也可以在创建任务的时候指定
	var taskId taskIDOption = "123"
	client := NewClient(RedisClientOpt{Addr: "localhost:6379"})
	t1 := NewTask("测试任务ID冲突1", nil, taskId)
	taskInfo, err := client.EnqueueContext(context.Background(), t1)
	if err != nil {
		t.Error(err)
	}
	t.Log(client, taskInfo)
	t2 := NewTask("测试任务ID冲突2", nil, taskId)
	taskInfo, err = client.EnqueueContext(context.Background(), t2)
	if err != nil {
		t.Error(err)
	}
	t.Log(client, taskInfo)
}
