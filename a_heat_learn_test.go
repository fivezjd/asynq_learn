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
// 3、心跳服务作为主服务的一部分，需要自己独立完成一些任务，所以可以在心跳结构体中定义一些心跳需要的东西，这些东西需要在初始化主服务的时候就 ToDo
// 定义
// 4、某些功能只需要执行一次，但是可能会有多个地方同时调用，所以可以使用sync.once ToDo
// 5、高密度的Redis命令可以使用lua脚本来执行 ToDo
// eg.
// 模拟心跳服务、模拟zset的使用、使用once、使用lua

func TestName(t *testing.T) {
	s := &ServerH{}
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
	Heat *HeatS
}

type HeatS struct {
	broker *redis.Client
}

func (a *ServerH) Run() {
	a.Heat = &HeatS{broker: redis.NewClient(&redis.Options{Addr: "localhost:6379"})}
	var wg sync.WaitGroup
	// 三种方式结束上下文
	// ctx, cancel := context.WithCancel(context.Background())
	// ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*20))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	a.Heat.Run(ctx, &wg)
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
	s.broker.ZAdd(context.Background(), strconv.Itoa(os.Getpid()), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: 1,
	})
	fmt.Println(time.Now().Unix())
	fmt.Println(strconv.Itoa(os.Getpid()))
}
