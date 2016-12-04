package main

import (
	"log"

	"time"

	"github.com/yuce/gobar"
)

type BarDateTime struct {
	format   string
	interval int
	info     gobar.BarSlotInfo
}

func (slot *BarDateTime) InitSlot(config map[string]interface{}, logger *log.Logger) (gobar.BarSlotConfig, error) {
	var info gobar.BarSlotInfo
	gobar.MapToBarSlotInfo(config, &info)
	slot.info = info
	if interval, ok := config["interval"].(int); ok {
		slot.interval = interval
	} else {
		slot.interval = 60
	}
	if format, ok := config["format"].(string); ok {
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
		slot.info.FullText = now.Format(slot.format)
		m := gobar.UpdateChannelMsg{
			ID:   ID,
			Info: slot.info,
		}
		updateChannel <- m
		time.Sleep(time.Duration(slot.interval) * time.Second)
	}
}
