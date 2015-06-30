package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

const (
	renderHost = "ws://renderhost.apps.zefr.com:4444/"
	origin     = "http://client.obsremote.com"
	feedSource = "http://monsterkill.apps.zefr.com:5000/action_feeds"
)

var (
	msgID = 2
)

type setOrderRequest struct {
	RequestType string   `json:"request-type"`
	SceneNames  []string `json:"scene-names"`
	MessageID   int      `json:"message-id,string"`
}

func nullWebsocket(ws *websocket.Conn) {
	go func() {
		var msg map[string]interface{}
		for {
			websocket.JSON.Receive(ws, &msg)
		}
	}()
}

func setNewActive(left int, right int, ws *websocket.Conn) error {
	order := make([]string, 0, 20)

	order = append(order, fmt.Sprintf("%03d-left", left))
	order = append(order, fmt.Sprintf("%03d-right", right))

	order = append(order, fmt.Sprintf("%03d-right", left))
	order = append(order, fmt.Sprintf("%03d-left", right))

	for i := 1; i < 11; i++ {
		if i != left && i != right {
			thisLeft := fmt.Sprintf("%03d-left", i)
			thisRight := fmt.Sprintf("%03d-right", i)
			order = append(order, thisLeft, thisRight)
		}
	}

	req := setOrderRequest{
		RequestType: "SetSourceOrder",
		SceneNames:  order,
		MessageID:   msgID,
	}
	msgID++
	return websocket.JSON.Send(ws, req)
}

type actionStatus struct {
	Players struct {
		Left struct {
			Name string `json:"hostName"`
		} `json:"blue"`
		Right struct {
			Name string `json:"hostName"`
		} `json:"red"`
	} `json:"feeds"`
}

func main() {
	ws, err := websocket.Dial(renderHost, "obsapi", origin)
	if err != nil {
		log.Fatal(err)
	}

	nullWebsocket(ws)

	err = setNewActive(5, 6, ws)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	for {
		resp, err := http.Get(feedSource)
		defer resp.Body.Close()

		if err != nil {
			log.Fatal(err)
		}

		current := new(actionStatus)

		decoder := json.NewDecoder(resp.Body)
		decoder.Decode(current)

		log.Print("left:" + current.Players.Left.Name)
		log.Print("right:" + current.Players.Right.Name)

		time.Sleep(1 * time.Second)
	}
}
