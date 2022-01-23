package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/miromax42/go-ftp-ci/internal/config"
	"github.com/miromax42/go-ftp-ci/internal/watcher"
)

const (
	dir        = "./test1"
	updateTime = time.Second
)

func main() {
	cfg, err := config.New("configs/ftpci.yml")
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("ftp host: %v \n", cfg.Ftp.Host)
	}

	w, err := watcher.New(cfg.Ftp)
	if err != nil {
		log.Fatalf("watcher.New: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dirs := []string{
		"/test1",
		"/test2",
	}

	notify := make(chan string, 5)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		go w.Watch(ctx, dirs, 5*time.Second, notify)

		<-ctx.Done()
		w.Stop()
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Executing stop")
				return
			case dir := <-notify:
				fmt.Printf("dir: %v, tasks: %v\n", dir, cfg.Tasks[dir])
				for _, c := range cfg.Tasks[dir] {
					cmd := exec.Command(c)
					cmd.Dir = "res"
					out, err := cmd.Output()
					if err != nil {
						fmt.Printf("cmd %s err: %v\n", c, err)
					} else {
						fmt.Printf("cmd %s out: %v\n", c, string(out))
					}

				}

			}
		}
	}()

	wg.Wait()
	fmt.Println("Gracefully done")
}
