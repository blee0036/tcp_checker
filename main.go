package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PingResult struct {
	Success bool  `json:"success"`
	Data    *Data `json:"data"`
}

type Data struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Ping string `json:"ping,omitempty"`
	Loss string `json:"loss,omitempty"`
}

type PingJob struct {
	host    string
	port    string
	results chan *TcpResult
}

type TcpResult struct {
	success bool
	ping    float64
}

var (
	attempts       int
	token          string
	timeoutMS      int
	pingJobChannel chan PingJob
)

func main() {
	var port int
	var threadNum int
	flag.IntVar(&port, "p", 8080, "port to listen on")
	flag.IntVar(&attempts, "a", 5, "number of connection attempts")
	flag.IntVar(&threadNum, "tN", 32, "number of tcp connect thread")
	flag.IntVar(&timeoutMS, "tO", 2000, "tcp connect timeout in milliseconds")
	flag.StringVar(&token, "t", "", "authentication token")
	flag.Parse()

	if port < 0 || port > 65535 {
		fmt.Println("Invalid port number. Please enter a port number between 0 and 65535.")
		return
	}

	if threadNum <= 0 {
		fmt.Println("Invalid thread number. Please enter a thread num more than 0.")
		return
	}

	if attempts <= 0 {
		fmt.Println("Invalid attempts number. Please enter attempts more than 0.")
		return
	}

	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/batch", handleBatchRequest)

	// start thread
	pingJobChannel = make(chan PingJob)
	for t := 0; t < threadNum; t++ {
		go tcpingThread(pingJobChannel)
	}

	fmt.Printf("HTTP server listening at http://0.0.0.0:%d\n", port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		fmt.Println("Start Http Server Error ", err.Error())
		return
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if token != "" {
		reqToken := r.URL.Query().Get("token")
		if reqToken != token {
			http.Error(w, "Forbidden - Invalid token", http.StatusForbidden)
			return
		}
	}
	host := r.URL.Query().Get("host")
	if host == "" {
		http.Error(w, "Missing host parameter", http.StatusBadRequest)
		return
	}

	port := r.URL.Query().Get("port")
	if port == "" {
		port = "80"
	} else {
		portInt, err := strconv.Atoi(port)
		if err != nil || portInt < 0 || portInt > 65535 {
			http.Error(w, "Invalid port number. Please enter a port number between 0 and 65535.", http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	result := performPing(host, port)

	jsonErr := json.NewEncoder(w).Encode(result)
	if jsonErr != nil {
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		return
	}
}

func handleBatchRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if token != "" {
		reqToken := r.URL.Query().Get("token")
		if reqToken != token {
			http.Error(w, "Forbidden - Invalid token", http.StatusForbidden)
			return
		}
	}

	var wg sync.WaitGroup
	results := make([]PingResult, 0)
	var mutex sync.Mutex

	scanner := bufio.NewScanner(r.Body)
	for scanner.Scan() {
		line := scanner.Text()
		hostPort := splitHostPort(line)
		if hostPort == nil {
			continue // skip invalid lines
		}

		wg.Add(1)
		go func(host, port string) {
			defer wg.Done()
			defer mutex.Unlock()

			result := performPing(host, port)
			mutex.Lock()
			results = append(results, *result)
		}(hostPort[0], hostPort[1])
	}

	if err := scanner.Err(); err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	jsonErr := json.NewEncoder(w).Encode(results)
	if jsonErr != nil {
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		return
	}
}

func splitHostPort(hostPortStr string) []string {
	hostPort := strings.Split(hostPortStr, ":")
	if len(hostPort) != 2 {
		return []string{hostPortStr, "80"}
	}
	portInt, err := strconv.Atoi(hostPort[1])
	if err != nil || portInt < 0 || portInt > 65535 {
		return nil
	}
	return hostPort
}

func tcpingThread(jobs <-chan PingJob) {
	for job := range jobs {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(job.host, job.port), time.Duration(timeoutMS)*time.Millisecond)
		if err == nil {
			ping := float64(time.Since(start).Milliseconds())
			closeErr := conn.Close()
			if closeErr != nil {
				fmt.Println("connect close err ", closeErr)
			}
			job.results <- &TcpResult{
				success: true,
				ping:    ping,
			}
		} else {
			job.results <- &TcpResult{
				success: false,
				ping:    0,
			}
		}

	}
}

func performPing(host, port string) *PingResult {
	successCount := 0
	totalPing := float64(0)

	results := make(chan *TcpResult)
	for i := 0; i < attempts; i++ {
		pingJobChannel <- PingJob{
			host:    host,
			port:    port,
			results: results,
		}
	}

	for i := 0; i < attempts; i++ {
		resultP := <-results
		result := *resultP
		if result.success {
			successCount += 1
			totalPing += result.ping
		}
	}

	if successCount == 0 {
		result := &PingResult{
			Success: false,
			Data: &Data{
				Host: host,
				Port: port,
			},
		}
		return result
	}
	loss := (float64(attempts) - float64(successCount)) / float64(attempts) * 100
	avgPing := totalPing / float64(successCount)

	result := &PingResult{
		Success: true,
		Data: &Data{
			Host: host,
			Port: port,
			Ping: fmt.Sprintf("%.2f", avgPing),
			Loss: fmt.Sprintf("%.2f", loss),
		},
	}

	return result
}
