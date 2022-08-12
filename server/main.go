package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/gorilla/mux"
)

type Command struct {
	Command   string
	Arguments []string
}

type Options struct {
	ConcurrentCommands int
}

func worker(c Command, doneChan *chan struct{}, o *Options) {
	defer close(*doneChan)
	log.Printf("Starting: %s ", c.Command)
	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}
	downloadCommand := cmd.NewCmdOptions(cmdOptions, c.Command, c.Arguments...)
	downloadCommand.Start() // non-blocking
	for downloadCommand.Stdout != nil || downloadCommand.Stderr != nil {
		select {
		case line, open := <-downloadCommand.Stdout:
			if !open {
				downloadCommand.Stdout = nil
				continue
			}
			log.Printf("%s", line)
		case line, open := <-downloadCommand.Stderr:
			if !open {
				downloadCommand.Stderr = nil
				continue
			}
			fmt.Fprintln(os.Stderr, line)
		}
	}

	log.Printf("Finished: %s %v", c.Command, c.Arguments)
}
func createDownloadHandler(o *Options) http.HandlerFunc {
	fn := func(w http.ResponseWriter, req *http.Request) {
		var c Command
		err := json.NewDecoder(req.Body).Decode(&c)
		doneChan := make(chan struct{})
		if err != nil {
			log.Println(err)
		}
		go worker(c, &doneChan, o)
		fmt.Fprintf(w, "Ran Command: %+v ", c)
	}
	return fn
}

func main() {
	concurrentCommands := flag.Int("c", 0, "Number of commands to run concurrently. 0 for unlimited.")
	flag.Parse()
	options := &Options{ConcurrentCommands: *concurrentCommands}
	r := newRouter(options)

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8090",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Starting up..")

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("Shutting down..")
	os.Exit(0)
}

func newRouter(o *Options) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", createDownloadHandler(o))
	return r
}