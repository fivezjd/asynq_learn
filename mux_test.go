package asynq_learn

import (
	"fmt"
	"github.com/hibiken/asynq/internal/errors"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// 以下代码演示了mux惯用的套路和配置对象参数的套路 from asynq_learn
var (
	defaultName = "zjd"
	defaultPort = 8080
)

type Config1 struct {
	Name string
	Port int
}

type Opt interface {
	I() int
	S() string
}

type setName string
type setPort int

func (s setPort) I() int {
	return int(s)
}

func (s setPort) S() string {
	return strconv.Itoa(int(s))
}

func (s setName) I() int {
	v, _ := strconv.Atoi(string(s))
	return v
}

func (s setName) S() string {
	return fmt.Sprintf("%s", string(s))
}

type handleF func(r io.Reader) error

type ServerMux struct {
	mu      sync.RWMutex       // 控制mu方法的原子性
	handMap map[string]handleF // 保存各个handle方法
	config  Config1
}

func NewServerMux(opts ...Opt) *ServerMux {
	ser := &ServerMux{
		handMap: make(map[string]handleF),
		config: Config1{
			Name: defaultName,
			Port: defaultPort,
		},
	}
	for _, opt := range opts {
		switch opt.(type) {
		case setName:
			ser.config.Name = opt.S()
		case setPort:
			ser.config.Port = opt.I()
		}
		// 其他类似
	}
	return ser
}

func (s *ServerMux) HandlerFunc(pattern string, hand handleF) error {
	if pattern == "" {
		return errors.New("parten is Empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handMap[pattern] = hand
	return nil
}

func (s *ServerMux) Run() {
	var wg sync.WaitGroup
	for _, f := range s.handMap {
		f := f
		wg.Add(1)
		go func() {
			err := f(strings.NewReader("这个字符串实现了io.reader"))
			if err != nil {
				log.Fatal(err)
			}
			wg.Done()
		}()
	}
	fmt.Println(s.config.Name)
	fmt.Println(s.config.Port)
	wg.Wait()
}

func TestMux(t *testing.T) {
	// 演示 修改配置的模式和注册mux注册handler的模式
	mu := NewServerMux(setPort(1024), setName("张建东"))
	err := mu.HandlerFunc("do", func(r io.Reader) error {
		// 读取10个字节
		s := make([]byte, 1024)
		read, err := r.Read(s)
		if err != nil {
			return err
		}
		fmt.Println(read)
		fmt.Println(string(s[:read]))
		fmt.Println("string(s)")
		return nil
	})
	if err != nil {
		return
	}
	mu.Run()
}
