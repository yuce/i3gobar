package main

import (
	"log"

	"time"

	"github.com/yuce/gobar"
)

type BarDateTime struct {
	textColor string
	format    string
}

func (slot *BarDateTime) InitSlot(config map[string]interface{}, logger *log.Logger) (gobar.BarSlotConfig, error) {
	if textColor, ok := config["text_color"].(string); ok {
		slot.textColor = textColor
	}
	if format, ok := config["text_color"].(string); ok {
		slot.format = format
	} else {
		slot.format = "2006-01-02 15:04:05"
	}
	return gobar.BarSlotConfig{
		MaxWidth:       -1,
		UpdateInterval: 0,
	}, nil
}

func (slot *BarDateTime) Start(ID int, updateChannel chan<- gobar.UpdateChannelMsg) {
	for {
		now := time.Now()
		m := gobar.UpdateChannelMsg{
			ID: ID,
			Info: gobar.BarSlotInfo{
				FullText:  now.Format(slot.format),
				TextColor: slot.textColor,
			},
		}
		updateChannel <- m
		time.Sleep(1 * time.Second)
	}
}
