package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

type PingResult struct {
	Success bool  `json:"success"`
	Data    *Data `json:"data,omitempty"`
}

type Data struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Ping string `json:"ping"`
	Loss string `json:"loss"`
}

var (
	attempts int
	token    string
)

func main() {
	var port int
	flag.IntVar(&port, "p", 8080, "port to listen on")
	flag.IntVar(&attempts, "a", 5, "number of connection attempts")
	flag.StringVar(&token, "t", "", "authentication token")
	flag.Parse()

	if port < 0 || port > 65535 {
		fmt.Println("Invalid port number. Please enter a port number between 0 and 65535.")
		return
	}

	if attempts <= 0 {
		fmt.Println("Invalid attempts number. Please enter attempts more than 0.")
		return
	}

	http.HandleFunc("/", handleRequest)
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

	successCount := 0
	totalPing := float64(0)

	for i := 0; i < attempts; i++ {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), time.Second*2)
		if err == nil {
			successCount++
			totalPing += float64(time.Since(start).Milliseconds())
			closeErr := conn.Close()
			if closeErr != nil {
				continue
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if successCount == 0 {
		result := &PingResult{
			Success: false,
		}
		jsonErr := json.NewEncoder(w).Encode(result)
		if jsonErr != nil {
			http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
			return
		}
		return
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

	jsonErr := json.NewEncoder(w).Encode(result)
	if jsonErr != nil {
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		return
	}
}
