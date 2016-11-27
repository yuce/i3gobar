package gobar

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"
)

// BarHeader i3 bar header
type barHeader struct {
	Version        int  `json:"version"`
	ClickEvents    bool `json:"click_events"`
	StopSignal     int  `json:"stop_signal"`
	ContinueSignal int  `json:"cont_signal"`
}

type BarSlotAlign string

const (
	CENTER BarSlotAlign = "center"
	RIGHT  BarSlotAlign = "right"
	LEFT   BarSlotAlign = "left"
)

type BarSlotMarkup string

const (
	NONE  BarSlotMarkup = "none"
	PANGO BarSlotMarkup = "pango"
)

type BarSlotInfo struct {
	FullText            string        `json:"full_text"`
	ShortText           string        `json:"short_text,omitempty"`
	TextColor           string        `json:"color,omitempty"`
	BackgroundColor     string        `json:"bacgkround,omitempty"`
	BorderColor         string        `json:"border,omitempty"`
	MinWidth            int           `json:"min_width,omitempty"`
	Align               BarSlotAlign  `json:"align,omitempty"`
	Name                string        `json:"name"`
	Instance            string        `json:"instance,omitempty"`
	IsUrgent            bool          `json:"urgent,omitempty"`
	HasSeparator        bool          `json:"separator,omitempty"`
	SeparatorBlockWidth int           `json:"separator_block_width,omitempty"`
	Markup              BarSlotMarkup `json:"markup,omitempty"`
}

type BarSlotConfig struct {
	MaxWidth       int
	UpdateInterval int64
}

type BarSlotConfigItem map[string]interface{}

// BarItem i3 bar item
type BarItem struct {
	// text          string
	// textWidth     int
	// scrollStartAt int
	Name       string
	InstanceOf string
	Slot       BarSlot
	SlotConfig map[string]interface{}
	info       *BarSlotInfo
	lastUpdate int64
	config     BarSlotConfig
}

type Bar struct {
	items  []BarItem
	logger *log.Logger
}

type BarSlot interface {
	InitSlot(config map[string]interface{}, logger *log.Logger) (BarSlotConfig, error)
	UpdateSlot() *BarSlotInfo
	// PauseSlot()
	// ResumeSlot()
}

func PrintHeader() {
	header := barHeader{
		Version:        1,
		ClickEvents:    false,
		StopSignal:     20, // SIGHUP
		ContinueSignal: 19, // SIGCONT
	}
	headerJSON, _ := json.Marshal(header)
	fmt.Println(string(headerJSON))
}

func CreateBar(items []BarItem, logger *log.Logger) *Bar {
	barItems := make([]BarItem, 0, len(items))
	var config BarSlotConfig
	var err error
	// var slot BarSlot
	now := time.Now().Unix()
	for _, item := range items {
		config, err = item.Slot.InitSlot(item.SlotConfig, logger)
		if err == nil {
			updateItem(&item, now)
			item.info.Name = item.Name
			item.info.Instance = item.InstanceOf
			item.config = config
			// item.Slot = slot
			barItems = append(barItems, item)
		} else {
			logger.Printf("Error: %q", err)
		}
	}
	return &Bar{
		items:  barItems,
		logger: logger,
	}
}

func updateItem(item *BarItem, now int64) {
	item.info = item.Slot.UpdateSlot()
	item.lastUpdate = now
}

func (bar *Bar) Update() {
	now := time.Now().Unix()
	for i, item := range bar.items {
		if item.config.UpdateInterval < 1 || now-item.lastUpdate < item.config.UpdateInterval {
			continue
		} else {
			bar.items[i].info = item.Slot.UpdateSlot()
			bar.items[i].lastUpdate = now
		}
	}
}

func (bar *Bar) Println() {
	var j []byte
	var err error
	bar.logger.Printf("%d %q", len(bar.items), bar.items)
	fmt.Println("[[]")
	for _, item := range bar.items {
		if item.info == nil {
			bar.logger.Printf("WARNING: slot info is nil: `%s`", item.Name)
			continue
		}
		j, err = json.Marshal(item.info)
		if err == nil {
			fmt.Printf(",%s\n", j)
		} else if j != nil {
			bar.logger.Printf("ERROR: %q", err)
		}
	}
	fmt.Println("]")
}

// func (bar *Bar) Marshal() string {
// 	j, err := json.Marshal(bar.items)
// 	if err != nil {
// 		return ""
// 	}
// 	return string(j)
// }

// func scrollText(text string, at int, textWidth int) (int, string) {
// 	if textWidth < 0 || at < 0 || len(text) < textWidth {
// 		return 0, text
// 	}
// 	if textWidth == 0 {
// 		return 0, ""
// 	}
// 	return (at + 1) % len(text), (text + " " + text)[at : at+textWidth]
// }

var typeRegistry = make(map[string]reflect.Type)

func AddModule(name string, typeOf reflect.Type) {
	typeRegistry[name] = typeOf
}

func createSlot(name string) (BarSlot, error) {
	v := reflect.New(typeRegistry[name]).Elem()
	fmt.Println("V", v, v.Interface())
	if slot, ok := v.Interface().(BarSlot); ok {
		return slot, nil
	}
	return nil, fmt.Errorf("Cannot create instance of `%s`", name)
}