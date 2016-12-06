package main

import (
	"log"

	"github.com/yuce/i3gobar"
)

type BarStaticText struct {
	info gobar.BarSlotInfo
}

func (slot *BarStaticText) InitSlot(config map[string]interface{}, barConfig *gobar.Configuration, logger *log.Logger) (gobar.BarSlotConfig, error) {
	info := gobar.MapToBarSlotInfo(config, barConfig)
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
