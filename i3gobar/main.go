package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yuce/i3gobar"
)

const (
	Stopped = iota
	Paused
	Running
)

type Configuration struct {
	Defaults    map[string]interface{}   `json:"defaults"`
	Items       []map[string]interface{} `json:"items"`
	barItems    []gobar.BarItem
	barDefaults gobar.BarSlotInfo
}

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
			// default:
			// 	runtime.Gosched()
			// 	if state == Paused {
			// 		break
			// 	}
		}
	}
}

func loadConfig(path string) (*Configuration, error) {
	text, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Configuration
	json.Unmarshal(text, &config)
	for _, ci := range config.Items {
		var item gobar.BarItem
		gobar.MapToBarItem(ci, &item)
		item.SlotConfig = ci
		config.barItems = append(config.barItems, item)
	}
	config.barDefaults = gobar.MapToBarSlotInfo(config.Defaults, nil)

	return &config, err
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
	logger.Printf("Loading configuration from: %s", configPath)
	config, err := loadConfig(configPath)
	barItems := config.barItems
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

	bar := gobar.CreateBar(barItems, &config.barDefaults, logger)
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
