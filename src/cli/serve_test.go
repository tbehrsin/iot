package main

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if err := os.Chdir("simpleapp"); err != nil {
		panic(err)
	}
	fmt.Println(m.Run())
}

func BenchmarkServeApp(b *testing.B) {
	var config *Config
	var err error
	if config, err = LoadConfig(); err != nil {
		panic(err)
	}

	gateway := FindGateway(config.Gateway)
	if gateway == nil {
		b.Errorf("unable to find gateway, check config file")
	}
	gateway.Token = config.Token
	gateway.Port = 443

	wg := &sync.WaitGroup{}
	wg.Add(b.N)
	fs := make([]*FileServer, b.N)
	fmt.Printf("testing with %d instances\n", b.N)
	for i := 0; i < b.N; i++ {
		if conn, err := gateway.WebSocket(); err != nil {
			b.Error(err)
		} else {
			dir, _ := os.Getwd()
			fs[i] = Serve(dir, conn)
			go func(fs *FileServer) {
				if err := fs.Wait(); err != nil {
					b.Error(err)
				}
				wg.Done()
			}(fs[i])
		}
	}

	c := make(chan bool)
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
	case <-time.After(4 * time.Second):
		for _, s := range fs {
			s.Close()
		}
	}
}
