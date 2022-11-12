package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func handler1(w http.ResponseWriter, req *http.Request) {
	fmt.Println("handler1 handling")
	time.Sleep(time.Duration(rand.Float64() * float64(time.Second)))
	statusCodes := []int{200, 400, 500}
	statusCode := 0
	r := rand.Intn(100)
	if r < 91 {
		statusCode = statusCodes[0]
	} else if r < 97 {
		statusCode = statusCodes[1]
	} else {
		statusCode = statusCodes[2]
	}
	w.WriteHeader(statusCode)
}

func handler2(w http.ResponseWriter, req *http.Request) {
	fmt.Println("handler2 handling")
	time.Sleep(time.Duration(rand.Float64() * float64(time.Second)))
	statusCodes := []int{200, 400, 500}
	statusCode := 0
	r := rand.Intn(100)
	if r < 93 {
		statusCode = statusCodes[0]
	} else if r < 95 {
		statusCode = statusCodes[1]
	} else {
		statusCode = statusCodes[2]
	}
	w.WriteHeader(statusCode)
}

func Run() {
	reg := prometheus.NewRegistry()
	metrics := NewHTTPMetrics(reg, 1, 0.2)
	metrics.StartBackgroundReportGoCollector(10 * time.Second)
	metrics.StartMetricsPush(10 * time.Second)
	http.HandleFunc("/handler1", metrics.WrapHandler("/handler1", handler1))
	http.HandleFunc("/handler2", metrics.WrapHandler("/handler2", handler2))
	http.HandleFunc("/metrics", metrics.MetricsHandler().ServeHTTP)

	http.ListenAndServe(":8083", nil)
}

func main() {
	Run()
}
