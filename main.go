package main

import (
	"embed"
	"flag"
	"github.com/gorilla/websocket"
	"io/fs"
	"log"
	"net/http"
	"time"
)

var (
	//go:embed static
	resources embed.FS
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func reader(ws *websocket.Conn) {
	defer ws.Close()
	for {
		_, message, err := ws.ReadMessage()
		log.Printf("Message: %s", string(message))

		if err != nil {
			log.Printf("Read Error: %s", err.Error())
			return
		}
	}
}

func writer(ws *websocket.Conn) {
	loop := time.NewTicker(time.Millisecond * 16)
	defer func() {
		loop.Stop()
		ws.Close()
	}()

	for {
		select {
		case <-loop.C:
			now := time.Now().Format("15:04:05.00000")
			if err := ws.WriteMessage(websocket.TextMessage, []byte(now)); err != nil {
				log.Printf("Write error: %s", err.Error())
				return
			}
		}
	}
}

func serveWebsocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	go writer(ws)
	reader(ws)
}

func main() {
	devMode := flag.Bool("d", false, "Enable dev mode.")
	flag.Parse()

	var files http.Handler
	if *devMode {
		files = http.FileServer(http.Dir("./static"))
	} else {
		// The fs.Sub line removes the static folder to match how devMode makes paths.
		serverRoot, _ := fs.Sub(resources, "static")
		files = http.FileServer(http.FS(serverRoot))
	}

	http.Handle("/", files)
	http.HandleFunc("/ws", serveWebsocket)

	if err := http.ListenAndServe(":8088", nil); err != nil {
		log.Fatal(err)
	}
}
