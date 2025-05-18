package main

import (
	"fmt"
	"github.com/iokiris/glockmon"
	"math/rand"
	"sync"
	"time"
)

var (
	mu       sync.Mutex
	workers  = make(map[int]chan struct{})
	workerID int
	wg       sync.WaitGroup
)

func StartWorkerManager(stopMain chan struct{}, muA, muB, muC *glockmon.MonitoredMutex) {
	go func() {
		tickerAdd := time.NewTicker(1 * time.Second)
		tickerRemove := time.NewTicker(3 * time.Second)
		const maxWorkers = 5

		for {
			select {
			case <-stopMain:
				tickerAdd.Stop()
				tickerRemove.Stop()
				return

			case <-tickerAdd.C:
				mu.Lock()
				if len(workers) < maxWorkers {
					workerID++
					id := workerID
					stopCh := make(chan struct{})

					var target *glockmon.MonitoredMutex
					var category string
					switch id % 3 {
					case 0:
						target, category = muA, "Category-A"
					case 1:
						target, category = muB, "Category-B"
					default:
						target, category = muC, "Category-C"
					}

					wg.Add(1)
					workers[id] = stopCh
					go startWorker(id, stopCh, target, category)
					fmt.Printf("[+] Started worker %d (%s). Total workers: %d\n", id, category, len(workers))
				}
				mu.Unlock()

			case <-tickerRemove.C:
				mu.Lock()
				for id, ch := range workers {
					close(ch)
					delete(workers, id)
					fmt.Printf("[-] Stopped worker %d. Total workers: %d\n", id, len(workers))
					break
				}
				mu.Unlock()
			}
		}
	}()
}

func startWorker(id int, stop <-chan struct{}, lock *glockmon.MonitoredMutex, category string) {
	defer wg.Done()
	for {
		select {
		case <-stop:
			fmt.Printf("[Worker %d] stopped (%s)\n", id, category)
			return
		default:
		}

		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

		lock.Lock()
		time.Sleep(800*time.Millisecond + time.Duration(rand.Intn(300))*time.Millisecond)
		lock.Unlock()
	}
}

func WaitForWorkers() {
	wg.Wait()
}
