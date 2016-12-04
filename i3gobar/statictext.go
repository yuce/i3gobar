package main

import (
	"log"

	"github.com/yuce/gobar"
)

type BarStaticText struct {
	text      string
	textColor string
}

func (slot *BarStaticText) InitSlot(config map[string]interface{}, logger *log.Logger) (gobar.BarSlotConfig, error) {
	if text, ok := config["text"].(string); ok {
		slot.text = text
	}
	if textColor, ok := config["text_color"].(string); ok {
		slot.textColor = textColor
	}
	return gobar.BarSlotConfig{
		MaxWidth:       -1,
		UpdateInterval: 0,
	}, nil
}

/*
func (slot BarStaticText) InitSlot(config map[string]interface{}, logger *log.Logger) (gobar.BarSlotConfig, error) {
	if text, ok := config["text"].(string); ok {
		slot.text = text
	}
	if textColor, ok := config["text_color"].(string); ok {
		slot.textColor = textColor
	}
	return gobar.BarSlotConfig{
		MaxWidth:       -1,
		UpdateInterval: 0,
	}, nil
}
*/

func (slot BarStaticText) Start(ID int, updateChannel chan<- gobar.UpdateChannelMsg) {
	m := gobar.UpdateChannelMsg{
		ID: ID,
		Info: gobar.BarSlotInfo{
			FullText:  slot.text,
			TextColor: slot.textColor,
		},
	}
	updateChannel <- m
}
