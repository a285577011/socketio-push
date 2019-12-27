package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

func doSomethingWith(c *gosocketio.Client, wg *sync.WaitGroup) {
	if res, err := c.Ack("join", "This is a client", time.Second*3); err != nil {
		log.Printf("error: %v", err)
	} else {
		log.Printf("result %q", res)
	}
	wg.Done()
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	for i:=0;i<20000;i++ {
		go func() {
			c, err := gosocketio.Dial(
				gosocketio.GetUrl("192.168.66.166", 5050, false),
				transport.GetDefaultWebsocketTransport())
			if err != nil {
				log.Fatal(err)
			}

			err = c.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
				log.Fatal("Disconnected")
			})
			if err != nil {
				log.Fatal(err)
			}

			err = c.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
				log.Println("Connected")
			})
			err = c.On("task", func(h *gosocketio.Channel) {
				log.Println("all",h.Id())
			})
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-quit
	log.Println("Shutdown Server ...")
}