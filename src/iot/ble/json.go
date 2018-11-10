package ble

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/paypal/gatt"
)

type JSONRequest struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

type JSONHandler func(req map[string]interface{}) (map[string]interface{}, error)

var JSONHandlers = make(map[string]JSONHandler)

func HandleFunc(t string, h JSONHandler) {
	JSONHandlers[t] = h
}

type JSONTransaction struct {
	Request  []byte
	Response []byte
}

type JSONConnection struct {
	Transactions  []JSONTransaction
	NotifyChannel chan struct{}
}

type JSONService struct {
	Service  *gatt.Service
	Centrals map[gatt.Central]*JSONConnection
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// reads chunks of messages on the queue
// packed data:
// struct {
//   id    uint8
// }
// followed by req.Central.MTU() - 1 bytes of binary data (JSON fragments)
// an id followed by no more data signifies the end of the message
// client keeps reading data until an empty response is sent
func (j *JSONService) Read(res gatt.ResponseWriter, req *gatt.ReadRequest) {
	c := j.Centrals[req.Central]

	for i, _ := range c.Transactions {
		t := &c.Transactions[i]
		if t.Response != nil && len(t.Response) > 0 {
			buffer := []byte{byte(i)}
			data := t.Response[:min(req.Central.MTU()-4, len(t.Response))]
			//log.Println("read: ", string(data))
			buffer = append(buffer, data...)

			res.SetStatus(gatt.StatusSuccess)
			if sz, err := res.Write(buffer); err != nil {
				log.Println(err)
				return
			} else if len(t.Response) > sz-1 {
				t.Response = t.Response[sz-1:]
			} else {
				t.Response = make([]byte, 0, 4)
			}
			return
		} else if t.Response != nil && len(t.Response) == 0 {
			buffer := []byte{byte(i)}
			res.Write(buffer)
			t.Response = nil
			return
		}
	}

	res.SetStatus(gatt.StatusSuccess)
	if _, err := res.Write([]byte{}); err != nil {
		log.Println(err)
	}
}

// r.Central.MTU()
// r.Central.Close()
//
// write message
// struct {
//   id uint8
// }
// followed by req.Central.MTU() - 1 bytes of binary data (JSON fragments)
// an id followed by no more data signifies the end of the message
func (j *JSONService) Write(r gatt.Request, data []byte) (status byte) {
	c := j.Centrals[r.Central]
	i := uint8(data[0])

	if c.Transactions[i].Request == nil {
		c.Transactions[i].Request = make([]byte, 0, r.Central.MTU())
		c.Transactions[i].Response = nil
	}
	//log.Println("write: ", string(c.Transactions[i].Request))
	if len(data) == 1 {
		// parse request into json
		var request JSONRequest
		if err := json.Unmarshal(c.Transactions[i].Request, &request); err != nil {
			log.Println(err)
			c.Transactions[i].Request = nil
			return
		}

		log.Printf("written: %+v\n", request)

		t := &c.Transactions[i]

		// delete request data
		t.Request = nil

		// call jsonhandler matching type with payload as goroutine
		if h, ok := JSONHandlers[request.Type]; !ok {
			out := make(map[string]interface{})
			out["error"] = fmt.Sprintf("Unknown message type \"%s\"", request.Type)
			var err error
			if t.Response, err = json.Marshal(out); err != nil {
				t.Response = []byte(`{"error": "Error encountered serializing error to json"}`)
			}
			c.NotifyChannel <- struct{}{}
		} else {
			go func(t *JSONTransaction, r map[string]interface{}, h JSONHandler) {
				// add response to transaction
				if response, err := h(r); err != nil {
					log.Println(r, err)
					out := make(map[string]interface{})
					out["error"] = err.Error()
					if t.Response, err = json.Marshal(out); err != nil {
						t.Response = []byte(`{"error": "Error encountered serializing error to json"}`)
					}
				} else {
					out := make(map[string]interface{})
					out["response"] = response
					// marshal response to json
					if d, err := json.Marshal(out); err != nil {
						log.Println(r, out, err)
						t.Response = []byte(`{"error": "Error encountered serializing response to json"}`)
					} else {
						t.Response = d
					}
				}

				log.Println("sending response: ", string(t.Response))

				// notify channel that there are new messages to read
				c.NotifyChannel <- struct{}{}
			}(t, request.Payload, h)
		}
	} else {
		c.Transactions[i].Request = append(c.Transactions[i].Request, data[1:]...)
	}

	return gatt.StatusSuccess
}

// notify of data available to read. empty message
//
// listen to channel and notify when there is a new message to read
//
func (j *JSONService) Notify(r gatt.Request, n gatt.Notifier) {
	c := j.Centrals[r.Central]

	for !n.Done() {
		<-c.NotifyChannel
		n.Write([]byte{})
	}
}

// Creates a JSONService object
func NewJSONService() *JSONService {
	j := &JSONService{
		Service:  gatt.NewService(gatt.MustParseUUID("1ed8da28-c0f8-42ee-90ed-e37619821619")),
		Centrals: make(map[gatt.Central]*JSONConnection),
	}

	c := j.Service.AddCharacteristic(gatt.MustParseUUID("706ff41f-e8d7-4b3b-88c0-c024ca7a41f7"))
	c.HandleReadFunc(j.Read)
	c.HandleWriteFunc(j.Write)
	c.HandleNotifyFunc(j.Notify)

	return j
}
