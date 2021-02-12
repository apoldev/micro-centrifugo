package centrifugo

import (
	"encoding/json"
	"log"

	"github.com/centrifugal/centrifuge-go"
)

type ResponseWriter struct {
	client      *centrifuge.Client
	Data        []byte
	userID      string
	initChannel string
}

type Response struct {
	Microservice string      `json:"microservice"`
	Payload      interface{} `json:"payload"`
}

func (r *ResponseWriter) Send(payload interface{}) {

	response := Response{
		Microservice: r.initChannel,
		Payload:      payload,
	}

	data, err := json.Marshal(response)

	result, err := r.client.Publish("#"+r.userID, data)

	if err != nil {
		log.Printf("Send to personal user channel #%s error:", r.initChannel, err)
		return
	}

	log.Printf("Send to personal user channel #%s result: %s", r.initChannel, result)

}
