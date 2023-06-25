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
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Only commands in this list will be allowed to be executed.
// TODO: Load from file dynamically.
var whiteList = []string{"echo"}

type Command struct {
	Command   string
	Arguments []string
}

type Options struct {
	ConcurrentCommands int
}
type Job struct {
	Command   Command
	StartTime time.Time
	EndTime   time.Time
	Status    string
	Ctx       context.Context    `json:"-"`
	Cancel    context.CancelFunc `json:"-"`
}

type Status struct {
	QueuedCommands   []Job
	RunningCommands map[string]Job
	CompletedCommands []Job
}

func worker(c Command, doneChan *chan struct{}, o *Options, s *Status, jobId string, ctx context.Context, cancel context.CancelFunc) {
	defer close(*doneChan)
	s.RunningCommands[jobId] = Job{Command: c, StartTime: time.Now(), Status: "Running", Ctx: ctx, Cancel: cancel}
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
			fmt.Fprintln(os.Stdout, line)
		case line, open := <-downloadCommand.Stderr:
			if !open {
				downloadCommand.Stderr = nil
				continue
			}
			fmt.Fprintln(os.Stderr, line)
		case <-ctx.Done():
			log.Printf("Canceling: %s ", c.Command)
			downloadCommand.Stop()
			if entry, ok := s.RunningCommands[jobId]; ok {
				entry.Status = "Cancelled"
				entry.EndTime = time.Now()
				s.RunningCommands[jobId] = entry
				return
			}
			if len(s.QueuedCommands) > 0 {
				job := s.QueuedCommands[0]
				s.QueuedCommands = s.QueuedCommands[1:]
				doneChan := make(chan struct{})
				jobId := uuid.New().String()
				ctx, cancel := context.WithCancel(context.Background())
				go worker(job.Command, &doneChan, o, s, jobId, ctx, cancel)
			}
		}
	}
	if entry, ok := s.RunningCommands[jobId]; ok {
		entry.Status = "Done"
		entry.EndTime = time.Now()
		// Delete from running commands and add to completed commands
		delete(s.RunningCommands, jobId)
		s.CompletedCommands = append(s.CompletedCommands, entry)
	}
	//Take the next job from the queue
	if len(s.QueuedCommands) > 0 {
		job := s.QueuedCommands[0]
		s.QueuedCommands = s.QueuedCommands[1:]
		doneChan := make(chan struct{})
		jobId := uuid.New().String()
		ctx, cancel := context.WithCancel(context.Background())
		go worker(job.Command, &doneChan, o, s, jobId, ctx, cancel)
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

func createDownloadHandler(o *Options, s *Status) http.HandlerFunc {
	fn := func(w http.ResponseWriter, req *http.Request) {
		enableCors(&w)
		log.Printf("Request: %s %s\n", req.RemoteAddr, req.Host)

		if (*req).Method == "OPTIONS" {
			return
		}
		var c Command
		err := json.NewDecoder(req.Body).Decode(&c)
		if err != nil {
			log.Println("Error decoding JSON: ", err)
			return
		}

		if !isCommandAllowed(c.Command) {
			log.Println("Command not allowed: ", c.Command)
			return
		}
		runningCommands := 0
		for _, job := range s.RunningCommands {
			if job.Status == "Running" {
				runningCommands++
			}
		}
		if o.ConcurrentCommands > 0 && runningCommands >= o.ConcurrentCommands {
			// Print number of currently running commands in a nice green color.
			log.Printf("\033[32mCurrently running %d commands, queuing %s\033[0m\n", runningCommands, c.Command)
			// Print number of commands already in Qeueu in a nice yellow color.
			log.Printf("\033[33mThere are now %d commands in queue.\033[0m\n", len(s.QueuedCommands) + 1)

			s.QueuedCommands = append(s.QueuedCommands, Job{Command: c, StartTime: time.Now(), Status: "Queued"})
			return
		}

		doneChan := make(chan struct{})
		jobId := uuid.New().String()
		ctx, cancel := context.WithCancel(context.Background())
		go worker(c, &doneChan, o, s, jobId, ctx, cancel)
		reponse := map[string]string{"jobId": jobId}
		json.NewEncoder(w).Encode(reponse)
	}
	return fn
}

func createStatusHandler(o *Options, s *Status) http.HandlerFunc {
	fn := func(w http.ResponseWriter, req *http.Request) {
		enableCors(&w)
		log.Printf("Request: %s %s\n", req.RemoteAddr, req.Host)

		if (*req).Method == "OPTIONS" {
			return
		}
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(s)
	}
	return fn
}

func createCancelHandler(o *Options, s *Status) http.HandlerFunc {
	fn := func(w http.ResponseWriter, req *http.Request) {
		enableCors(&w)
		log.Printf("Request: %s %s\n", req.RemoteAddr, req.Host)

		if (*req).Method == "OPTIONS" {
			return
		}
		vars := mux.Vars(req)
		jobId := vars["jobId"]
		if entry, ok := s.RunningCommands[jobId]; ok {
			fmt.Println("calling cancel")
			entry.Cancel()
		}

		w.Header().Set("Content-Type", "application/json")
		responseStatus := map[string]string{"status": "ok"}
		json.NewEncoder(w).Encode(responseStatus)
	}
	return fn
}

func main() {
	concurrentCommands := flag.Int("c", 1, "Number of commands to run concurrently. 0 for unlimited.")
	flag.Parse()
	options := &Options{ConcurrentCommands: *concurrentCommands}
	status := &Status{RunningCommands: make(map[string]Job)}
	r := newRouter(options, status)

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

func newRouter(o *Options, s *Status) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", createDownloadHandler(o, s))
	r.HandleFunc("/status", createStatusHandler(o, s))
	r.HandleFunc("/{jobId}/cancel", createCancelHandler(o, s))
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
