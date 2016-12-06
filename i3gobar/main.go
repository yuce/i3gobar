package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"path/filepath"

	"github.com/yuce/i3gobar"
)

const (
	Stopped = iota
	Paused
	Running
)

type Configuration struct {
	Defaults map[string]interface{}   `json:"defaults"`
	Items    []map[string]interface{} `json:"items"`
	Theme    string                   `json:"theme"`
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
		default:
			runtime.Gosched()
			if state == Paused {
				break
			}
			bar.Update()
		}
	}
}

func loadTheme(path string) (gobar.Theme, error) {
	barTheme := gobar.Theme{}
	text, err := ioutil.ReadFile(path)
	if err != nil {
		return barTheme, err
	}
	json.Unmarshal(text, &barTheme)
	return barTheme, nil
}

func loadConfig(path string, logger *log.Logger) (barConfig gobar.Configuration, err error) {
	barConfig = gobar.Configuration{}
	text, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	var config Configuration
	json.Unmarshal(text, &config)
	logger.Printf("Theme: %s", config.Theme)
	if config.Theme != "" {
		abspath, err := filepath.Abs(path)
		abspath = filepath.Dir(abspath)
		var themePath string
		if err == nil {
			themePath = filepath.Join(abspath, config.Theme)
		} else {
			logger.Printf("Trouble getting absolute path: %q", err)
			themePath = config.Theme
		}
		logger.Printf("Loading theme: %s", themePath)
		theme, err := loadTheme(themePath)
		if err != nil {
			logger.Printf("Error while loading theme: %q", err)
		}
		barConfig.Theme = &theme
	}
	for _, ci := range config.Items {
		var item gobar.BarItem
		gobar.MapToBarItem(ci, &item)
		item.SlotConfig = ci
		barConfig.Items = append(barConfig.Items, item)
	}
	defaults := gobar.MapToBarSlotInfo(config.Defaults, nil)
	barConfig.Defaults = &defaults

	return
}

func checkErr(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s:%q", msg, err)
		os.Exit(2)
	}
}

func main() {
	var logPath, configPath string
	flag.StringVar(&logPath, "log", "/dev/null", "Log path. Default: /dev/null")
	flag.StringVar(&logPath, "l", "/dev/null", "Log file to use. Default: /dev/null")
	flag.StringVar(&configPath, "config", "", "Configuration path.")
	flag.StringVar(&configPath, "c", "", "Configuration path (in JSON).")

	flag.Parse()

	logfile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	checkErr(err, "Log file not opened")
	defer logfile.Close()
	logger := log.New(logfile, "gobar:", log.Lshortfile|log.LstdFlags)

	Init()
	logger.Printf("Loading configuration from: %s", configPath)
	config, err := loadConfig(configPath, logger)
	checkErr(err, "Config file not opened")
	logger.Printf("bar items: %q", config.Items)
	items := config.Items
	for i := 0; i < len(items); i++ {
		slot, err := gobar.CreateSlot(items[i].Module)
		if err != nil {
			items[i].Slot = &BarStaticText{}
			items[i].Label = "ERR: "
			items[i].SlotConfig = map[string]interface{}{
				"text_color": "#FF0000",
				"full_text":  err.Error(),
			}
		} else {
			items[i].Slot = slot
		}
	}

	bar := gobar.CreateBar(&config, logger)
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
