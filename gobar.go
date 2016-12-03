package gobar

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
)

// BarHeader i3 bar header
type barHeader struct {
	Version     int  `json:"version"`
	ClickEvents bool `json:"click_events"`
	// StopSignal     int  `json:"stop_signal"`
	// ContinueSignal int  `json:"cont_signal"`
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
	ID         int
	Name       string
	InstanceOf string
	Slot       BarSlot
	SlotConfig map[string]interface{}
	info       BarSlotInfo
	lastUpdate int64
	config     BarSlotConfig
}

type UpdateChannelMsg struct {
	ID   int
	Info BarSlotInfo
}

type Bar struct {
	items         []BarItem
	logger        *log.Logger
	updateChannel chan UpdateChannelMsg
}

type BarSlot interface {
	InitSlot(config map[string]interface{}, logger *log.Logger) (BarSlotConfig, error)
	Start(ID int, updateChannel chan<- UpdateChannelMsg)
	// PauseSlot()
	// ResumeSlot()
}

func PrintHeader() {
	header := barHeader{
		Version:     1,
		ClickEvents: false,
		// StopSignal:     20, // SIGHUP
		// ContinueSignal: 19, // SIGCONT
	}
	headerJSON, _ := json.Marshal(header)
	fmt.Println(string(headerJSON))
	fmt.Println("[[]")
}

func CreateBar(items []BarItem, logger *log.Logger) *Bar {
	barItems := make([]BarItem, 0, len(items))
	var config BarSlotConfig
	var err error
	updateChannel := make(chan UpdateChannelMsg)
	for _, item := range items {
		config, err = item.Slot.InitSlot(item.SlotConfig, logger)
		if err == nil {
			// updateItem(&item, now)
			item.info.Name = item.Name
			item.info.Instance = item.InstanceOf
			item.config = config
			barItems = append(barItems, item)
			go item.Slot.Start(len(barItems)-1, updateChannel)
		} else {
			logger.Printf("Error: %q", err)
		}
	}
	return &Bar{
		items:         barItems,
		logger:        logger,
		updateChannel: updateChannel,
	}
}

func (bar *Bar) Update() {
	for {
		select {
		case m := <-bar.updateChannel:
			bar.items[m.ID].info = m.Info
			bar.Println()

		}
	}
}

func (bar *Bar) Println() {
	var j []byte
	var err error
	bar.logger.Printf("%d %q", len(bar.items), bar.items)
	fmt.Printf(",[\n")
	for i, item := range bar.items {
		j, err = json.Marshal(item.info)
		if err == nil {
			if i == 0 {
				fmt.Printf("%s\n", j)
			} else {
				fmt.Printf(",%s\n", j)
			}
		} else if j != nil {
			bar.logger.Printf("ERROR: %q", err)
		}
	}
	fmt.Println("]")
}

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
