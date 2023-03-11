// Copyright 2020 Kentaro Hibino. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package asynq_learn_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/hibiken/asynq"
	"golang.org/x/sys/unix"
)

func ExampleServer_Run() {
	srv := asynq_learn.NewServer(
		asynq_learn.RedisClientOpt{Addr: ":6379"},
		asynq_learn.Config{Concurrency: 20},
	)

	h := asynq_learn.NewServeMux()
	// ... Register handlers

	// Run blocks and waits for os signal to terminate the program.
	if err := srv.Run(h); err != nil {
		log.Fatal(err)
	}
}

func ExampleServer_Shutdown() {
	srv := asynq_learn.NewServer(
		asynq_learn.RedisClientOpt{Addr: ":6379"},
		asynq_learn.Config{Concurrency: 20},
	)

	h := asynq_learn.NewServeMux()
	// ... Register handlers

	if err := srv.Start(h); err != nil {
		log.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, unix.SIGTERM, unix.SIGINT)
	<-sigs // wait for termination signal

	srv.Shutdown()
}

func ExampleServer_Stop() {
	srv := asynq_learn.NewServer(
		asynq_learn.RedisClientOpt{Addr: ":6379"},
		asynq_learn.Config{Concurrency: 20},
	)

	h := asynq_learn.NewServeMux()
	// ... Register handlers

	if err := srv.Start(h); err != nil {
		log.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, unix.SIGTERM, unix.SIGINT, unix.SIGTSTP)
	// Handle SIGTERM, SIGINT to exit the program.
	// Handle SIGTSTP to stop processing new tasks.
	for {
		s := <-sigs
		if s == unix.SIGTSTP {
			srv.Stop() // stop processing new tasks
			continue
		}
		break // received SIGTERM or SIGINT signal
	}

	srv.Shutdown()
}

func ExampleScheduler() {
	scheduler := asynq_learn.NewScheduler(
		asynq_learn.RedisClientOpt{Addr: ":6379"},
		&asynq_learn.SchedulerOpts{Location: time.Local},
	)

	if _, err := scheduler.Register("* * * * *", asynq_learn.NewTask("task1", nil)); err != nil {
		log.Fatal(err)
	}
	if _, err := scheduler.Register("@every 30s", asynq_learn.NewTask("task2", nil)); err != nil {
		log.Fatal(err)
	}

	// Run blocks and waits for os signal to terminate the program.
	if err := scheduler.Run(); err != nil {
		log.Fatal(err)
	}
}

func ExampleParseRedisURI() {
	rconn, err := asynq_learn.ParseRedisURI("redis://localhost:6379/10")
	if err != nil {
		log.Fatal(err)
	}
	r, ok := rconn.(asynq_learn.RedisClientOpt)
	if !ok {
		log.Fatal("unexpected type")
	}
	fmt.Println(r.Addr)
	fmt.Println(r.DB)
	// Output:
	// localhost:6379
	// 10
}

func ExampleResultWriter() {
	// ResultWriter is only accessible in Handler.
	h := func(ctx context.Context, task *asynq_learn.Task) error {
		// .. do task processing work

		res := []byte("task result data")
		n, err := task.ResultWriter().Write(res) // implements io.Writer
		if err != nil {
			return fmt.Errorf("failed to write task result: %v", err)
		}
		log.Printf(" %d bytes written", n)
		return nil
	}

	_ = h
}
