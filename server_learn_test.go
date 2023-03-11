package asynq_learn

import (
	"context"
	"strings"
	"testing"
)

func TestClient(t *testing.T) {
	client := NewClient(RedisClientOpt{
		Addr: "localhost:6379",
	})
	t1 := NewTask("task1", []byte{1, 2}, groupOption("group-t"))
	t2 := NewTask("task2", []byte{3, 4}, groupOption("group-t"))
	client.Enqueue(t1)
	client.Enqueue(t2)
}

func TestMaxSize(t *testing.T) {
	server := NewServer(RedisClientOpt{Addr: "localhost:6379"}, Config{
		GroupMaxSize: 1000000,
		GroupAggregator: GroupAggregatorFunc(func(group string, tasks []*Task) *Task {
			t.Log(group)
			var b strings.Builder
			for _, task := range tasks {
				b.Write(task.payload)
				b.WriteString("\n")
			}
			return NewTask("aggregated-task", []byte(b.String()))
		}),
	})
	mux := NewServeMux()
	mux.HandleFunc("aggregated-task", func(ctx context.Context, task *Task) error {
		t.Log(string(task.payload))
		return nil
	})
	server.Run(mux)
}
