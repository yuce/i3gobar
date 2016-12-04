package main

import (
	"log"

	"time"

	"os/exec"

	"strings"

	"github.com/yuce/gobar"
)

type BarExternalCommand struct {
	command  string
	interval int
	info     gobar.BarSlotInfo
}

func (slot *BarExternalCommand) InitSlot(config map[string]interface{}, defaults *gobar.BarSlotInfo, logger *log.Logger) (gobar.BarSlotConfig, error) {
	info := gobar.MapToBarSlotInfo(config, defaults)
	slot.info = info
	if command, ok := config["command"].(string); ok {
		slot.command = command
	}
	if interval, ok := config["interval"].(float64); ok {
		slot.interval = int(interval)
	} else {
		slot.interval = 0
	}
	return gobar.BarSlotConfig{
		MaxWidth:       -1,
		UpdateInterval: 0,
	}, nil
}

func (slot *BarExternalCommand) Start(ID int, updateChannel chan<- gobar.UpdateChannelMsg) {
	if slot.command == "" {
		slot.info.FullText = "ERROR: Missing command"
		slot.info.TextColor = "#FF0000"
		updateChannel <- gobar.UpdateChannelMsg{
			ID:   ID,
			Info: slot.info,
		}
		return
	}

	for {
		out, err := exec.Command("sh", "-c", slot.command).Output()
		if err != nil {
			slot.info.FullText = err.Error()
			slot.info.TextColor = "#FF2222"
			slot.interval = 0
		} else {
			slot.info.FullText = strings.TrimSpace(string(out))
		}
		m := gobar.UpdateChannelMsg{
			ID:   ID,
			Info: slot.info,
		}
		updateChannel <- m
		if slot.interval > 0 {
			time.Sleep(time.Duration(slot.interval) * time.Second)
		} else {
			return
		}

	}
}
