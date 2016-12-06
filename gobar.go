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
	AlignCenter BarSlotAlign = "center"
	AlignRight  BarSlotAlign = "right"
	AlignLeft   BarSlotAlign = "left"
)

type BarSlotMarkup string

const (
	MarkupNone  BarSlotMarkup = "none"
	MarkupPango BarSlotMarkup = "pango"
)

type BarSlotInfo struct {
	FullText            string        `json:"full_text"`
	ShortText           string        `json:"short_text,omitempty"`
	TextColor           string        `json:"color,omitempty"`
	BackgroundColor     string        `json:"background,omitempty"`
	BorderColor         string        `json:"border,omitempty"`
	MinWidth            int           `json:"min_width,omitempty"`
	Align               BarSlotAlign  `json:"align,omitempty"`
	Name                string        `json:"name"`
	Instance            string        `json:"instance,omitempty"`
	IsUrgent            bool          `json:"urgent"`
	HasSeparator        bool          `json:"separator"`
	SeparatorBlockWidth int           `json:"separator_block_width"`
	Markup              BarSlotMarkup `json:"markup,omitempty"`
}

type BarSlotConfig struct {
	MaxWidth       int
	UpdateInterval int64
}

type BarSlotConfigItem map[string]interface{}

// BarItem i3 bar item
type BarItem struct {
	ID         int
	Name       string `json:"name"`
	Module     string `json:"module"`
	Label      string `json:"label"`
	Theme      string `json:"theme"`
	Slot       BarSlot
	SlotConfig map[string]interface{} //`json:"config"`
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

type ThemeItem struct {
	Color      string `json:"color"`
	Background string `json:"background"`
	Border     string `json:"border"`
}

type Theme struct {
	Items map[string]ThemeItem `json:"items"`
}

type Configuration struct {
	Defaults *BarSlotInfo
	Theme    *Theme
	Items    []BarItem
}

type BarSlot interface {
	InitSlot(config map[string]interface{}, barConfig *Configuration, logger *log.Logger) (BarSlotConfig, error)
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

func CreateBar(barConfig *Configuration, logger *log.Logger) *Bar {
	var err error
	// var barItems []BarItem
	updateChannel := make(chan UpdateChannelMsg)
	items := barConfig.Items
	var item BarItem
	for i := 0; i < len(items); i++ {
		item = items[i]
		_, err = item.Slot.InitSlot(item.SlotConfig, barConfig, logger)
		if err == nil {
			item.info.Name = item.Name
			item.info.Instance = item.Module
			// item.config = config
			go item.Slot.Start(i, updateChannel)
		} else {
			logger.Printf("Error: %q", err)
		}
	}
	return &Bar{
		items:         items,
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
		item.info.FullText = fmt.Sprintf(" %s%s ", item.Label, item.info.FullText)
		item.info.Name = item.Name
		item.info.Instance = item.Module

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

func CreateSlot(name string) (BarSlot, error) {
	if entry, ok := typeRegistry[name]; ok {
		v := reflect.New(entry)
		if slot, ok := v.Interface().(BarSlot); ok {
			return slot, nil
		}
		return nil, fmt.Errorf("Cannot create instance of `%s`", name)
	}
	return nil, fmt.Errorf("Module not found: `%s`", name)
}

func MapToBarSlotInfo(m map[string]interface{}, barConfig *Configuration) BarSlotInfo {
	b, err := json.Marshal(m)
	var info BarSlotInfo
	if barConfig != nil && barConfig.Defaults != nil {
		info = *barConfig.Defaults
	} else {
		info = BarSlotInfo{
			MinWidth:            0,
			Align:               AlignLeft,
			IsUrgent:            false,
			HasSeparator:        true,
			SeparatorBlockWidth: 10,
		}
	}
	if themeKey, ok := m["theme"]; ok && barConfig.Theme != nil {
		if themeKeyString, ok := themeKey.(string); ok {
			if themeItem, ok := barConfig.Theme.Items[themeKeyString]; ok {
				if themeItem.Color != "" {
					info.TextColor = themeItem.Color
				}
				if themeItem.Background != "" {
					info.BackgroundColor = themeItem.Background
				}
				if themeItem.Border != "" {
					info.BorderColor = themeItem.Border
				}
			}
		}
	}
	if err != nil {
		return info
	}
	json.Unmarshal(b, &info)
	return info
}

func MapToBarItem(m map[string]interface{}, item *BarItem) {
	b, err := json.Marshal(m)
	if err != nil {
		item = &BarItem{}
		return
	}
	json.Unmarshal(b, item)
}
