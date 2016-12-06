package main

import (
	"log"
	"time"

	"github.com/yuce/i3gobar"
)

type BarDateTime struct {
	format   string
	location *time.Location
	interval int
	info     gobar.BarSlotInfo
}

func (slot *BarDateTime) InitSlot(config map[string]interface{}, defaults *gobar.BarSlotInfo, logger *log.Logger) (gobar.BarSlotConfig, error) {
	info := gobar.MapToBarSlotInfo(config, defaults)
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
	if location, ok := config["location"].(string); ok {
		zone, err := time.LoadLocation(location)
		if err != nil {
			logger.Printf("Timezone not found: `%s", location)
		} else {
			slot.location = zone
		}
	}
	return gobar.BarSlotConfig{
		MaxWidth:       -1,
		UpdateInterval: 0,
	}, nil
}

func (slot *BarDateTime) Start(ID int, updateChannel chan<- gobar.UpdateChannelMsg) {
	var now time.Time
	for {
		now = time.Now()
		if slot.location != nil {
			now = now.In(slot.location)
		}

		slot.info.FullText = now.Format(slot.format)
		m := gobar.UpdateChannelMsg{
			ID:   ID,
			Info: slot.info,
		}
		updateChannel <- m
		time.Sleep(time.Duration(slot.interval) * time.Second)
	}
}
