package jst

import (
	"syscall/js"
)

// func FuncOf(fn func(this Value, args []Value) interface{}) Func {

var Document = js.Global().Get("document")
var LocalStorage = js.Global().Get("localStorage")

func Id(id string) js.Value {
	return Document.Call("getElementById", id)
}
