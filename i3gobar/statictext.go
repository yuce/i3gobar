package main

import (
	"log"

	"github.com/yuce/gobar"
)

type BarStaticText struct {
	info gobar.BarSlotInfo
}

func (slot *BarStaticText) InitSlot(config map[string]interface{}, defaults *gobar.BarSlotInfo, logger *log.Logger) (gobar.BarSlotConfig, error) {
	info := gobar.MapToBarSlotInfo(config, defaults)
	slot.info = info

	return gobar.BarSlotConfig{
		MaxWidth:       -1,
		UpdateInterval: 0,
	}, nil
}

func (slot BarStaticText) Start(ID int, updateChannel chan<- gobar.UpdateChannelMsg) {
	m := gobar.UpdateChannelMsg{
		ID:   ID,
		Info: slot.info,
	}
	updateChannel <- m
}
