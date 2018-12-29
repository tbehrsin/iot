package apps

import (
	"log"
	"net/http"

	v8 "github.com/behrsin/go-v8"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Inspector struct {
	inspector *v8.Inspector
	write     chan string
}

func NewInspector() *Inspector {
	i := &Inspector{
		write: make(chan string),
	}

	go func() {
		log.Fatal(http.ListenAndServe(":9222", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println("inspector upgrade:", err)
				return
			}

			defer func() {
				c.Close()
				close(i.write)
				i.write = make(chan string)
			}()

			go func() {
				for {
					message, more := <-i.write

					if more {
						// log.Println("RX>", string(message))
						c.WriteMessage(websocket.TextMessage, []byte(message))
					} else {
						break
					}
				}
			}()

			for {
				if mt, message, err := c.ReadMessage(); err != nil {
					return
				} else if mt == websocket.TextMessage {
					if i.inspector != nil {
						// log.Println("TX>", string(message))
						i.inspector.DispatchMessage(string(message))
					}
				} else if mt == websocket.CloseMessage {
					return
				}
			}
		})))
	}()

	return i
}

func (i *Inspector) V8InspectorSendResponse(callID int, message string) {
	if i.write != nil {
		i.write <- message
	}
}

func (i *Inspector) V8InspectorSendNotification(message string) {
	if i.write != nil {
		i.write <- message
	}
}

func (i *Inspector) V8InspectorFlushProtocolNotifications() {

}

func (i *Inspector) AddApp(app *App) {
	log.Println("Inspector.AddApp")
	defer log.Println("Inspector.AddApp exit")
	i.inspector = app.isolate.NewInspector(i)
	i.inspector.AddContext(app.context, app.Package().Name())
}

func (i *Inspector) RemoveApp(app *App) {
	log.Println("Inspector.RemoveApp")
	defer log.Println("Inspector.RemoveApp exit")
	i.inspector.RemoveContext(app.context)
	i.inspector.Release()
	i.inspector = nil
}
