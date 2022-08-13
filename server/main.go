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

// Only commands in this list will be allowed to be executed.
// TODO: Load from file dynamically.
var whiteList = []string{"echo", "cowsay"}

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

func isCommandAllowed(command string) bool {
	for _, c := range whiteList {
		if c == command {
			return true
		}
	}
	return false
}

func createDownloadHandler(o *Options) http.HandlerFunc {
	fn := func(w http.ResponseWriter, req *http.Request) {
		enableCors(&w)
		fmt.Printf("%s %s\n", req.RemoteAddr, req.Host)
		if (*req).Method == "OPTIONS" {
			return
		}
		var c Command
		err := json.NewDecoder(req.Body).Decode(&c)
		if err != nil {
			log.Println("Error decoding JSON: ", err)
		}

		if !isCommandAllowed(c.Command) {
			log.Println("Command not allowed: ", c.Command)
			return
		}

		doneChan := make(chan struct{})

		go worker(c, &doneChan, o)
		fmt.Fprintf(w, "Sent Command: %+v ", c)
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

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	(*w).Header().Set("Access-Control-Expose-Headers", "Content-Length")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
