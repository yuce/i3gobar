package main

import (
	"log"

	"reflect"

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

func (slot *BarStaticText) UpdateSlot() *gobar.BarSlotInfo {
	return &gobar.BarSlotInfo{
		FullText:  slot.text,
		TextColor: slot.textColor,
	}
}

// func (slot *BarStaticText) PauseSlot() {

// }

// func (slot *BarStaticText) ResumeSlot() {

// }

func Init() {
	gobar.AddModule("StaticText", reflect.TypeOf(BarStaticText{}))
}
