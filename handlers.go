package centrifugo

import (
	"log"

	"github.com/centrifugal/centrifuge-go"
)

type PublishWrapper struct {
	handler *OnPublishHandler
	sub     *centrifuge.Subscription
	e       centrifuge.PublishEvent
}

type OnPublishHandler struct {
	client  *centrifuge.Client
	handler func(*ResponseWriter)
}

func (h *OnPublishHandler) OnPublish(sub *centrifuge.Subscription, e centrifuge.PublishEvent) {

	log.Printf("User %s: to channel %s: %s", e.Info.User, sub.Channel(), string(e.Data))

	if h.handler != nil {
		publishWriter := &ResponseWriter{
			client:      h.client,
			Data:        e.Data,
			userID:      e.Info.User,
			initChannel: sub.Channel(),
		}
		h.handler(publishWriter)
	}

}
