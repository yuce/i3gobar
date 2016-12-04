package main

import (
	"log"

	"time"

	"os/exec"

	"strings"

	"github.com/yuce/gobar"
)

type BarExternalCommand struct {
	textColor string
	command   string
	interval  int
}

func (slot *BarExternalCommand) InitSlot(config map[string]interface{}, logger *log.Logger) (gobar.BarSlotConfig, error) {
	if textColor, ok := config["text_color"].(string); ok {
		slot.textColor = textColor
	}
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
		updateChannel <- gobar.UpdateChannelMsg{
			ID: ID,
			Info: gobar.BarSlotInfo{
				FullText:  "ERROR: Missing command",
				TextColor: "#FF0000",
			},
		}
		return
	}
	var text, textColor string
	for {
		out, err := exec.Command("sh", "-c", slot.command).Output()
		if err != nil {
			text = err.Error()
			textColor = "#FF2222"
			slot.interval = 0
		} else {
			text = strings.TrimSpace(string(out))
			textColor = slot.textColor
		}
		m := gobar.UpdateChannelMsg{
			ID: ID,
			Info: gobar.BarSlotInfo{
				FullText:  text,
				TextColor: textColor,
			},
		}
		updateChannel <- m
		if slot.interval > 0 {
			time.Sleep(time.Duration(slot.interval) * time.Second)
		} else {
			return
		}

	}
}
