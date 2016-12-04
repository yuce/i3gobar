package main

import (
	"reflect"

	"github.com/yuce/gobar"
)

func Init() {
	gobar.AddModule("StaticText", reflect.TypeOf(BarStaticText{}))
	gobar.AddModule("DateTime", reflect.TypeOf(BarDateTime{}))
	gobar.AddModule("ExternalCommand", reflect.TypeOf(BarExternalCommand{}))
}
