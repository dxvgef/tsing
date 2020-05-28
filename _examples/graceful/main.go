package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dxvgef/tsing"
)

var ignoreSignals = []os.Signal{os.Interrupt}
var forwardSignals = []os.Signal{syscall.SIGTERM}

func main() {
	mainCtx, cancelFunc := context.WithCancel(context.Background())
	handler := tsing.New(&tsing.Config{})
	handler.GET("/slow", func(c *tsing.Context) error {
		log.Println("do request start")
		time.Sleep(10 * time.Second)
		log.Println("do  request finished")
		c.ResponseWriter.Write([]byte("finished"))
		return nil
	})

	httpSvr := &http.Server{
		Addr:         ":5656",
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second}
	mainSig := make(chan os.Signal, 4)
	signal.Notify(mainSig, ignoreSignals...)
	signal.Notify(mainSig, forwardSignals...)
	shutdown := make(chan struct{}, 1)
	defer cancelFunc()
	go func() {
		if err := httpSvr.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("start http srv failed:%s", err)
			}
		}
		log.Println("closing")
		close(shutdown)
	}()

	go func() {
		<-mainSig
		log.Println("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := httpSvr.Shutdown(ctx); err != nil {
			log.Fatalf("close http srv failed:%s", err)
		}
		log.Println("shutdown ok")
		select {
		case <-shutdown:
		case <-ctx.Done():
		}
		log.Println("shutdown finished")
		cancelFunc()
	}()

	<-mainCtx.Done()
}
