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

func main() {
	logfile, err := os.OpenFile("/dev/null", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		fmt.Println("ERR:", err)
	}
	defer logfile.Close()
	logger := log.New(logfile, "gobar:", log.Lshortfile|log.LstdFlags)
	Init()
	barItems := make([]gobar.BarItem, 3)
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
	barItems[1] = gobar.BarItem{
		Name:       "text2",
		InstanceOf: "StaticText",
		Slot:       &BarStaticText{},
		SlotConfig: map[string]interface{}{
			"name":       "text2",
			"module":     "StaticText",
			"text_color": "#00FF00",
			"text":       "Yet another bar item!",
		},
	}
	barItems[2] = gobar.BarItem{
		Name:       "datetime1",
		InstanceOf: "DateTime",
		Slot:       &BarDateTime{},
		SlotConfig: map[string]interface{}{
			"name":   "datetime1",
			"module": "DateTime",
			// "text_color": "#0000FF",
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
	logger.Println("Ended")
}
