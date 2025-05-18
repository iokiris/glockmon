package main

import (
	"fmt"
	"github.com/iokiris/glockmon"
	"github.com/iokiris/glockmon/config"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Initialize the lock monitor with default config
	monitor := glockmon.NewMonitor(config.Default())

	// Create monitored mutexes with categories
	muA := glockmon.New(monitor, 50*time.Millisecond)
	muA.SetCategory("Category-A")

	muB := glockmon.New(monitor, 50*time.Millisecond)
	muB.SetCategory("Category-B")

	muC := glockmon.New(monitor, 50*time.Millisecond)
	muC.SetCategory("Category-C")

	// Start and manage workers (logic inside workers.go)
	stopMain := make(chan struct{})
	StartWorkerManager(stopMain, muA, muB, muC)

	// exit on Ctrl + C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Press Ctrl+C to stop...")
	<-sigs

	close(stopMain)
	WaitForWorkers()
	fmt.Println("All workers stopped. Exiting.")
}
