package main

import (
	"container/heap"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/saurabh/protohackers/internal/logger"
)

type BaseRequest struct {
	Request string `json:"request"`
}

type GetRequest struct {
	Request string   `json:"request"`
	Queues  []string `json:"queues"`
	Wait    bool     `json:"wait,omitempty"`
}

type PutRequest struct {
	Request string `json:"request"`
	Queue   string `json:"queue"`
	Job     any    `json:"job"`
	Pri     uint32 `json:"pri"`
}

type AbortRequest struct {
	Request string `json:"request"`
	Id      string `json:"id"`
}

type DeleteRequest struct {
	Request string `json:"request"`
	Id      string `json:"id"`
}

type JobServer struct {
	queue   map[string]*PriorityQueue // queue name -> priority queue
	mu      sync.RWMutex
	numJobs uint64
}

type JobItem struct {
	id    string
	job   any
	pri   uint32
	index int
}

type PriorityQueue []*JobItem

func (pq PriorityQueue) Len() int {
	return len(pq)
}

// We need a max heap
func (pq PriorityQueue) Less(i int, j int) bool {
	return pq[i].pri > pq[j].pri
}

func (pq PriorityQueue) Swap(i int, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*JobItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // don't stop the GC from reclaiming the item eventually
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Peek() (*JobItem, error) {
	if len(*pq) == 0 {
		return nil, fmt.Errorf("queue is empty")
	}
	return (*pq)[0], nil
}

func (js *JobServer) Put(queue string, job any, pri uint32) uint64 {
	js.mu.Lock()
	defer js.mu.Unlock()

	if _, exists := js.queue[queue]; !exists {
		js.queue[queue] = &PriorityQueue{}
		heap.Init(js.queue[queue])
	}
	js.numJobs++

	newJob := &JobItem{
		id:  strconv.FormatUint(js.numJobs, 10),
		job: job,
		pri: pri,
	}
	heap.Push(js.queue[queue], newJob)
	return js.numJobs
}

func (js *JobServer) Get(queues []string, wait bool) (string, error) {
	js.mu.Lock()
	defer js.mu.Unlock()

	bestQueue := ""
	bestJob := &JobItem{}
	bestPri := uint32(0)
	for _, queue := range queues {
		if _, exists := js.queue[queue]; !exists {
			log.Println("Queue not found:", queue)
			continue
		}
		job, err := js.queue[queue].Peek()
		if bestJob == nil || (err != nil && job.pri >= bestPri) {
			bestJob = job
			bestPri = job.pri
			bestQueue = queue
		}
	}
	if bestJob == nil {
		log.Println("No job found")
		return "", errors.New("no job found")
	}
	log.Println("Got job", bestJob.id, "from queue", bestQueue, "with priority", bestPri)
	return bestJob.id, nil
}

var port = flag.String("port", "50001", "Port to listen on")

func main() {
	flag.Parse()

	// Setup logging to logs directory
	logFile, err := logger.Setup("job-center")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	ln, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	log.Println("Listening on port " + *port)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		cancel()
		ln.Close()
	}()

	js := JobServer{
		queue:   make(map[string]*PriorityQueue),
		numJobs: uint64(0),
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				log.Println("Server stopped")
				return
			default:
				log.Println("Accept error:", err)
				continue
			}
		}
		go handleConnection(conn, &js)
	}
}

func handleConnection(conn net.Conn, js *JobServer) {
	log.Println("New connection from", conn.RemoteAddr())
	defer conn.Close()

	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)
	enc.SetEscapeHTML(false)

	for {

		// The server must not close the connection in response to an invalid request.
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			log.Println("Decode error:", err)
			_ = enc.Encode(map[string]any{"status": "error", "error": "Decode error: " + err.Error()})
			continue
		}
		var req BaseRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			log.Println("Unmarshal error:", err)
			_ = enc.Encode(map[string]any{"status": "error", "error": "Unmarshal error: " + err.Error()})
			continue
		}

		switch req.Request {
		case "get":
			log.Println("Get request:", string(raw))
			var getReq GetRequest
			if err := json.Unmarshal(raw, &getReq); err != nil {
				log.Println("Unmarshal error:", err)
				_ = enc.Encode(map[string]any{"status": "error", "error": "Unmarshal error: " + err.Error()})
				continue
			}
			jobId, _ := js.Get(getReq.Queues, getReq.Wait)
			if jobId != "" {
				_ = enc.Encode(map[string]any{"status": "ok", "id": jobId})
			} else {
				_ = enc.Encode(map[string]any{"status": "no-job"})
			}
		case "put":
			log.Println("Put request:", raw)
			var putReq PutRequest
			if err := json.Unmarshal(raw, &putReq); err != nil {
				log.Println("Unmarshal error:", err)
				_ = enc.Encode(map[string]any{"status": "error", "error": "Unmarshal error: " + err.Error()})
				continue
			}
			js.Put(putReq.Queue, putReq.Job, putReq.Pri)

		case "abort":
			log.Println("Abort request:", raw)
			var abortReq AbortRequest
			if err := json.Unmarshal(raw, &abortReq); err != nil {
				log.Println("Unmarshal error:", err)
				_ = enc.Encode(map[string]any{"status": "error", "error": "Unmarshal error: " + err.Error()})
				continue
			}
		case "delete":
			log.Println("Delete request:", raw)
			var deleteReq DeleteRequest
			if err := json.Unmarshal(raw, &deleteReq); err != nil {
				log.Println("Unmarshal error:", err)
				_ = enc.Encode(map[string]any{"status": "error", "error": "Unmarshal error: " + err.Error()})
				continue
			}
		default:
			log.Println("Invalid request:", raw)
			_ = enc.Encode(map[string]any{"status": "error", "error": "Unrecognised request type."})
		}
	}
}
