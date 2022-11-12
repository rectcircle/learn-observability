package main

import (
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func RequestHandler(handlerName string) {
	resp, err := http.Get("http://localhost:8083" + handlerName)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func TestRun(t *testing.T) {
	go Run()
	handlerNameChan := make(chan string)
	go func() {
		for {
			if rand.Float64() < 0.6 {
				handlerNameChan <- "/handler1"
			} else {
				handlerNameChan <- "/handler2"
			}
		}
	}()
	for i := 0; i < 100; i++ { // 并发度
		go func() {
			for {
				RequestHandler(<-handlerNameChan)
			}
		}()
	}
	time.Sleep(130 * time.Second)
}
