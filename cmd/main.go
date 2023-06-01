package main

import (
	"context"
	"fmt"
	co2_monitor "github.com/A-ndrey/co2-monitor"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	metricChan, err := co2_monitor.Start(ctx, 500*time.Millisecond)
	if err != nil {
		log.Fatalln(err)
	}
	defer cancelFunc()

	for metric := range metricChan {
		fmt.Println(metric)
	}
}
