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
	feedList   = "http://monsterkill.apps.zefr.com:5000/feeds_list"
)

var (
	msgID = 2
)

type setOrderRequest struct {
	RequestType string   `json:"request-type"`
	SceneNames  []string `json:"scene-names"`
	MessageID   int      `json:"message-id,string"`
}

type setSourceRender struct {
	RequestType string `json:"request-type"`
	Source      string `json:"source"`
	Render      bool   `json:"render"`
	MessageID   int    `json:"message-id,string"`
}

func nullWebsocket(ws *websocket.Conn) {
	go func() {
		var msg map[string]interface{}
		for {
			websocket.JSON.Receive(ws, &msg)
		}
	}()
}

var oldLeft, oldRight int

func setNewActive(left int, right int, ws *websocket.Conn) error {
	if oldLeft != left || oldRight != right {
		oldLeft, oldRight = left, right
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
		log.Println(order)

		req := setOrderRequest{
			RequestType: "SetSourceOrder",
			SceneNames:  order,
			MessageID:   msgID,
		}
		msgID++
		return websocket.JSON.Send(ws, req)
	}
	return nil
}

func setSide(num int, side string, ws *websocket.Conn) error {
	left := setSourceRender{
		RequestType: "SetSourceRender",
		Source:      fmt.Sprintf("%03d-left", num),
		MessageID:   msgID,
	}
	msgID++
	right := setSourceRender{
		RequestType: "SetSourceRender",
		Source:      fmt.Sprintf("%03d-right", num),
		MessageID:   msgID,
	}
	msgID++

	if side == "left" {
		left.Render = true
		right.Render = false
	} else if side == "right" {
		right.Render = true
		left.Render = false
	} else {
		log.Fatal("wtf did you give me?", side)
	}

	err := websocket.JSON.Send(ws, left)
	if err != nil {
		return err
	}
	websocket.JSON.Send(ws, right)
	if err != nil {
		return err
	}
	return nil
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

type feedsList struct {
	Players []struct {
		HostName string `json:"hostName"`
		Team     string `json:"team"`
	} `json:"players"`
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

sleepLoop:
	for {
		time.Sleep(1 * time.Second)
		resp, err := http.Get(feedSource)

		if err != nil {
			log.Print(err)
			continue sleepLoop
		}
		defer resp.Body.Close()

		current := new(actionStatus)

		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(current)
		if err != nil {
			log.Println(err)
			continue sleepLoop
		}

		var leftInt, rightInt int
		_, err = fmt.Sscanf(current.Players.Left.Name, "zefr%03d", &leftInt)
		if err != nil {
			log.Print(err)
			continue sleepLoop
		}
		_, err = fmt.Sscanf(current.Players.Right.Name, "zefr%03d", &rightInt)
		if err != nil {
			log.Print(err)
			continue sleepLoop
		}
		log.Printf("left: %d, right: %d", leftInt, rightInt)
		err = setNewActive(leftInt, rightInt, ws)
		if err != nil {
			log.Println(err)
		}

		resp, err = http.Get(feedList)
		if err != nil {
			log.Print(err)
			continue sleepLoop
		}
		defer resp.Body.Close()

		feeds := new(feedsList)
		decoder = json.NewDecoder(resp.Body)
		err = decoder.Decode(&feeds)
		if err != nil {
			log.Println(err)
			continue sleepLoop
		}

		for _, feed := range feeds.Players {
			var thisNum int
			_, err = fmt.Sscanf(feed.HostName, "zefr%03d", &thisNum)
			if err != nil {
				log.Print(err)
				continue sleepLoop
			}
			if feed.Team == "Blue" {
				err = setSide(thisNum, "left", ws)
			} else if feed.Team == "Red" {
				err = setSide(thisNum, "right", ws)
			} else {
				log.Print("wtf is this team", feeds)
				continue sleepLoop
			}
			if err != nil {
				log.Print(err)
				continue sleepLoop
			}
		}

	}
}
