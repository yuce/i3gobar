package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"os"

	"flag"

	"io/ioutil"

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

func loadConfig(path string) (items []gobar.BarItem, err error) {
	text, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var configItems []map[string]interface{}
	json.Unmarshal(text, &configItems)
	for _, ci := range configItems {
		var item gobar.BarItem
		gobar.MapToBarItem(ci, &item)
		// mapstructure.Decode(ci, &item)
		item.SlotConfig = ci
		items = append(items, item)
	}
	return
}

func main() {
	var logPath, configPath string
	flag.StringVar(&logPath, "log", "/dev/null", "Log path. Default: /dev/null")
	flag.StringVar(&logPath, "l", "/dev/null", "Log file to use. Default: /dev/null")
	flag.StringVar(&configPath, "config", "", "Configuration path.")
	flag.StringVar(&configPath, "c", "", "Configuration path (in JSON).")

	flag.Parse()

	logfile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		fmt.Println("ERR:", err)
	}
	defer logfile.Close()
	logger := log.New(logfile, "gobar:", log.Lshortfile|log.LstdFlags)

	Init()
	var barItems []gobar.BarItem
	logger.Printf("Loading configuration from: %s", configPath)
	barItems, err = loadConfig(configPath)
	if err != nil {
		barItems = make([]gobar.BarItem, 1)
		barItems[0] = gobar.BarItem{
			Name:  "text1",
			Label: "ERR: ",
			Slot:  &BarStaticText{},
			SlotConfig: map[string]interface{}{
				"text_color": "#FF0000",
				"full_text":  err.Error(),
			},
		}
	} else {
		for i := 0; i < len(barItems); i++ {
			slot, err := gobar.CreateSlot(barItems[i].Module)
			if err != nil {
				barItems[i].Slot = &BarStaticText{}
				barItems[i].Label = "ERR: "
				barItems[i].SlotConfig = map[string]interface{}{
					"text_color": "#FF0000",
					"full_text":  err.Error(),
				}
			} else {
				barItems[i].Slot = slot
			}
		}
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
