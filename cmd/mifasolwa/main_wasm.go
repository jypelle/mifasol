package main

import (
	"syscall/js"
	"time"
)

func main() {
	doc := js.Global().Get("document")
	body := doc.Get("body")
	body.Set("innerHTML", `<main>
    <h1>Welcome to mifasol</h1>
    <h2>Download your client</h2>
    <ul>
        <li><a href="/clients/mifasolcli-windows-amd64.exe">Windows (amd64)</a></li>
        <li><a href="/clients/mifasolcli-linux-amd64">Linux (amd64)</a></li>
        <li><a href="/clients/mifasolcli-linux-arm">Linux (arm)</a></li>
    </ul>
	<p id="message">...</p>
</main>`)

	message := doc.Call("getElementById", "message")
	time.Sleep(2 * time.Second)
	message.Set("innerHTML", "5")
	time.Sleep(2 * time.Second)
	message.Set("innerHTML", "4")
	time.Sleep(2 * time.Second)
	message.Set("innerHTML", "3")
	time.Sleep(2 * time.Second)
	message.Set("innerHTML", "2")
	time.Sleep(2 * time.Second)
	message.Set("innerHTML", "1")
	time.Sleep(2 * time.Second)
	message.Set("innerHTML", "0")
}
