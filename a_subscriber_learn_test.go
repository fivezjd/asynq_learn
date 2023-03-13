package asynq_learn

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"sync"
	"testing"
	"time"
)

// 实现一个程序
// 结构体S 有订阅程序p 有任务程序t ,订阅程序负责监听Redis一个键，然后在map中找到对应的取消程序，尽力去取消任务t

func TestNa(t *testing.T) {
	s := NewS()
	s.Run()
}

type S struct {
	p *P
	t *T
}

func NewS() *S {
	cals := &cMap{d: make(map[int]context.CancelFunc)}
	return &S{
		p: &P{
			cancelMap: cals,
			broker:    redis.NewClient(&redis.Options{Addr: "localhost:6379"}),
		},
		t: &T{
			cancelMap: cals,
		},
	}
}

func (s *S) Run() {
	var wg sync.WaitGroup
	s.p.Run(&wg)
	s.t.Run(&wg)
	wg.Wait()
}

type cMap struct {
	mu sync.Mutex
	d  map[int]context.CancelFunc
}

func (c *cMap) set(id int, f context.CancelFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.d[id]; ok {
		return errors.New("映射已存在")
	}
	c.d[id] = f
	return nil
}

func (c *cMap) del(id int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.d, id)
	fmt.Println("删除", id)
}

func (c *cMap) get(id int) (context.CancelFunc, error) {
	if f, ok := c.d[id]; ok {
		return f, nil
	}
	return nil, errors.New("映射不存在")
}

type P struct {
	cancelMap *cMap
	broker    *redis.Client
}

const KeyS = "subs_task"

func (p *P) Run(group *sync.WaitGroup) {
	group.Add(1)
	go func() {
		defer group.Done()
		var (
			subs *redis.PubSub
		)
		for {
			// 订阅Redis
			subs = p.broker.Subscribe(context.Background(), KeyS)
			_, err := subs.Receive(context.Background())
			if err != nil {
				continue
			}
			break
		}
		channel := subs.Channel()
		for {
			select {
			case msg := <-channel:
				id, _ := strconv.Atoi(msg.Payload)
				f, err := p.cancelMap.get(id)
				if err != nil {
					fmt.Println("任务不存在")
					continue
				}
				f()
				fmt.Println("收到客户端发布的消息，取消任务", id)
			}
		}
	}()
	// 订阅任务
}

type T struct {
	cancelMap *cMap
}

func (t *T) Run(group *sync.WaitGroup) {
	ctxSlice := make([]context.Context, 0)
	group.Add(1)
	// 添加任务
	go func() {
		defer group.Done()
		for i := 0; i < 10; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10+time.Second*time.Duration(i*20))
			t.cancelMap.d[i] = cancel
			ctxSlice = append(ctxSlice, ctx)
		}
		for i, ctx := range ctxSlice {
			ctx := ctx
			i := i
			go func() {
				select {
				case <-ctx.Done():
					t.cancelMap.del(i)
				}
			}()
		}

	}()
	select {}
}

//在 Go 语言中，在循环中生成 goroutine 可能会导致一些问题，主要有以下几点需要注意：
//
//闭包问题：由于 goroutine 是异步执行的，当在循环中使用闭包时，可能会出现意外的结果。具体来说，如果在循环中启动 goroutine 时，使用了循环变量作为闭包参数，那么在 goroutine 实际执行时，循环变量的值可能已经发生了变化，导致结果不符合预期。解决该问题的方法是将循环变量赋值给一个新的局部变量，并将该局部变量传递给闭包。
//
//goroutine 数量控制：在循环中启动大量的 goroutine 可能会导致系统负载过高，甚至导致程序崩溃。因此，应该对启动的 goroutine 数量进行控制，例如可以使用带缓冲的通道或者 WaitGroup 等方法来控制 goroutine 的数量。
//
//goroutine 资源释放：如果在循环中启动的 goroutine 涉及到一些资源（如文件句柄、数据库连接等），需要确保在 goroutine 执行完毕后及时释放这些资源，避免资源泄露导致程序崩溃或者系统性能下降。
//
//程序退出问题：在循环中启动的 goroutine 可能会持续运行，导致程序无法退出。为了避免这种情况，可以使用 context 包来控制 goroutine 的退出，或者使用 sync.WaitGroup 等方法等待所有 goroutine 执行完毕后再退出程序。
