package asynq_learn

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

// 通过此项目的心跳监测代码，有以下几点的发现
// 1、要用Redis实现一个队列的话，可以使用List类型，但是如果队列中的元素有不同的过期时间的话，就不能用List.
// 可以使用zset有序集合类型来实现队列 score 表示元素的过期时间，这样不仅能实现元素的有序性，还可以按score来淘汰元素
// 2、一般的服务中，为了定时的同步当前主程的状态，可以使用一个goroutine来实现，类似一个for循环中监听不同的通道，来做不同的监听任务。
// 一般的心跳是指定时间之后做，但是需要重复做这件事，所以可以用 time.Timer 类型，监听到timer之后可以使用reset方法来重置timer
// 3、心跳服务作为主服务的一部分，需要自己独立完成一些任务，所以可以在心跳结构体中定义一些心跳需要的东西，这些东西需要在初始化主服务的时候就
// 定义
// 4、某些功能只需要执行一次，但是可能会有多个地方同时调用，所以可以使用sync.once
// 5、高密度的Redis命令可以使用lua脚本来执行 保证原子性
// eg.
// 模拟心跳服务、模拟zset的使用、使用once、使用lua

func TestName(t *testing.T) {
	s := NewServerH()
	s.Run()
}

func Test1(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
		Password: "",               // no password set
		DB:       0,                // use default DB
	})

	pong, err := rdb.Ping(context.Background()).Result()
	fmt.Println(pong, err)
}

type ServerH struct {
	Heat        *HeatS
	HealthCheck *HealthCheck
	ServerA
}

type HealthCheck struct {
	once   sync.Once
	broker *redis.Client
	done   chan struct{}
}

func (hc *HealthCheck) check() {
	pong, err := hc.broker.Ping(context.Background()).Result()
	if err != nil {
		hc.once.Do(func() {
			fmt.Println("服务器异常")
			hc.done <- struct{}{}
		})
	}
	fmt.Println(pong)

}

func (hc *HealthCheck) Run(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		hc.check()
		internal := time.NewTimer(time.Second * 5)
		for {
			select {
			case <-internal.C:
				hc.check()
				internal.Reset(time.Second * 5)
				fmt.Println("5秒检查一次")
			case v := <-hc.done:
				fmt.Println(v)
				fmt.Println("结束健康监测")
				return
			}
		}
	}()
}

func NewServerH() *ServerH {
	return &ServerH{
		Heat: &HeatS{broker: redis.NewClient(&redis.Options{Addr: "localhost:6379"})},
		HealthCheck: &HealthCheck{
			broker: redis.NewClient(&redis.Options{Addr: "localhost:6379"}),
			done:   make(chan struct{}, 1),
		},
		ServerA: ServerA{log: NewLoggerA(NewBaseLog(os.Stdout))},
	}
}

type HeatS struct {
	broker *redis.Client
}

func (a *ServerH) Run() {
	var wg sync.WaitGroup
	// 三种方式结束上下文
	// ctx, cancel := context.WithCancel(context.Background())
	// ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*20))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	a.Heat.Run(ctx, &wg)
	a.HealthCheck.Run(&wg)
	// 一分钟之后结束心跳程序
	time.AfterFunc(time.Minute, func() {
		cancel()
	})
	wg.Wait()
}

func (s *HeatS) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.beat()
		t := time.NewTimer(time.Second * 5)
		for {
			select {
			// 监听通道
			case <-t.C:
				t.Reset(time.Second * 5)
				fmt.Println("timer is reset")
				s.beat()
				// 每隔5秒做一些事情
			case <-ctx.Done():
				fmt.Println("心跳监测结束")
				return
			}
		}
	}()
}

func (s *HeatS) beat() {
	// 实现Redis zset操作
	// 添加或者更新元素
	s.NX()
	s.XX()
	s.ZREM()

}

func (s *HeatS) ZREM() {
	luaScript := `
		return redis.call('ZREM', KEYS[1], ARGV[1])
`
	scriptSha := s.broker.ScriptLoad(context.Background(), luaScript).Val()
	// result 是移除的元素数量
	result, err := s.broker.EvalSha(context.Background(), scriptSha, []string{strconv.Itoa(os.Getpid())}, 1).Result()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result, strconv.Itoa(os.Getpid()))
}

// 只更新
func (s *HeatS) XX() {
	luaScript := `
		return redis.call('ZADD', KEYS[1], 'XX', ARGV[1], ARGV[2])
`
	script := redis.NewScript(luaScript)
	result, err := script.Run(context.Background(), s.broker, []string{strconv.Itoa(os.Getpid())}, 100, 1).Result()
	if err != nil {
		return
	}
	fmt.Println(result, strconv.Itoa(os.Getpid()))
}

// 只添加
func (s *HeatS) NX() {
	luaScript := `
		return redis.call('ZADD', KEYS[1], 'NX', ARGV[1], ARGV[2])
`
	script := redis.NewScript(luaScript)
	result, err := script.Run(context.Background(), s.broker, []string{strconv.Itoa(os.Getpid())}, 2, 1).Result()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result, strconv.Itoa(os.Getpid()))
}
