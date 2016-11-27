package main

import (
	"fmt"
	"log"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"os"

	"github.com/yuce/gobar"
)

const oneSec = 1000000000

const (
	Stopped = iota
	Paused
	Running
)

func updateLoop(bar *gobar.Bar, ws <-chan int, logger *log.Logger) {
	state := Running
	for {
		logger.Printf("State: %d", state)
		select {
		case state = <-ws:
			logger.Printf("Received new state: %d", state)
			if state == Stopped {
				return
			}
		default:
			runtime.Gosched()
			if state == Paused {
				break
			}
			logger.Println("Updating")
			bar.Update()
			bar.Println()
			time.Sleep(oneSec)
		}
	}
}

// slot, err = createSlot(item.InstanceOf)
// if err != nil {
// 	logger.Printf("Warning: Undefined module: `%s`: %q", item.InstanceOf, err)
// 	continue
// }

func main() {
	logfile, err := os.OpenFile("/home/yuce/ramdisk/gobar.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		fmt.Println("ERR:", err)
	}
	defer logfile.Close()
	logger := log.New(logfile, "gobar:", log.Lshortfile|log.LstdFlags)
	Init()
	barItems := make([]gobar.BarItem, 1)
	barItems[0] = gobar.BarItem{
		Name:       "text1",
		InstanceOf: "StaticText",
		Slot:       &BarStaticText{},
		SlotConfig: map[string]interface{}{
			"name":       "text1",
			"module":     "StaticText",
			"text_color": "#FF0000",
			"text":       "Hello, World!",
		},
	}
	bar := gobar.CreateBar(barItems, logger)
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	ws := make(chan int, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGHUP, syscall.SIGCONT)
	logger.Println("Started")
	go func() {
		gobar.PrintHeader()
		go updateLoop(bar, ws, logger)
		ws <- Running
	end:
		for {
			logger.Println("Waiting for a signal")
			sig := <-sigs
			logger.Printf("Received signal: %q", sig)
			switch sig {
			case syscall.SIGHUP:
				ws <- Paused
			case syscall.SIGCONT:
				ws <- Running
			default:
				break end
			}
		}
		done <- true
	}()
	<-done
	// bar := gobar.CreateBar()
	// updateLoop(bar, ws, logger)
	logger.Println("Ended")
}
