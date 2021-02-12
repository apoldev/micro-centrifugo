package centrifugo

import (
	"log"

	"github.com/centrifugal/centrifuge-go"
	"github.com/dgrijalva/jwt-go"
)

const exampleTokenHmacSecret = "my_secret"

func connToken(user string, exp int64) string {
	// NOTE that JWT must be generated on backend side of your application!
	// Here we are generating it on client side only for example simplicity.
	claims := jwt.MapClaims{"sub": user}
	if exp > 0 {
		claims["exp"] = exp
	}
	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(exampleTokenHmacSecret))
	if err != nil {
		panic(err)
	}
	return t
}

type eventHandler struct{}

type Client struct {
	client  *centrifuge.Client
	handler *eventHandler
}

func (h *eventHandler) OnConnect(_ *centrifuge.Client, e centrifuge.ConnectEvent) {
	log.Printf("Connected to micro with ID %s", e.ClientID)
}

func (h *eventHandler) OnError(_ *centrifuge.Client, e centrifuge.ErrorEvent) {
	log.Printf("Error: %s", e.Message)
}

func (h *eventHandler) OnMessage(_ *centrifuge.Client, e centrifuge.MessageEvent) {
	log.Printf("Message from server: %s", string(e.Data))
}

func (h *eventHandler) OnDisconnect(_ *centrifuge.Client, e centrifuge.DisconnectEvent) {
	log.Printf("Disconnected from micro: %s", e.Reason)
}

func (h *eventHandler) OnServerSubscribe(_ *centrifuge.Client, e centrifuge.ServerSubscribeEvent) {
	log.Printf("Subscribe to server-side channel %s: (resubscribe: %t, recovered: %t)", e.Channel, e.Resubscribed, e.Recovered)
}

func (h *eventHandler) OnServerUnsubscribe(_ *centrifuge.Client, e centrifuge.ServerUnsubscribeEvent) {
	log.Printf("Unsubscribe from server-side channel %s", e.Channel)
}

func (h *eventHandler) OnServerJoin(_ *centrifuge.Client, e centrifuge.ServerJoinEvent) {
	log.Printf("Server-side join to channel %s: %s (%s)", e.Channel, e.User, e.Client)
}

func (h *eventHandler) OnServerLeave(_ *centrifuge.Client, e centrifuge.ServerLeaveEvent) {
	log.Printf("Server-side leave from channel %s: %s (%s)", e.Channel, e.User, e.Client)
}

func (h *eventHandler) OnServerPublish(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	log.Printf("Publication from server-side channel %s: %s", e.Channel, e.Data)
}

func (h *eventHandler) OnJoin(sub *centrifuge.Subscription, e centrifuge.JoinEvent) {
	log.Printf("Someone joined %s: user id %s, client id %s", sub.Channel(), e.User, e.Client)
}

func (h *eventHandler) OnLeave(sub *centrifuge.Subscription, e centrifuge.LeaveEvent) {
	log.Printf("Someone left %s: user id %s, client id %s", sub.Channel(), e.User, e.Client)
}

func (h *eventHandler) OnSubscribeSuccess(sub *centrifuge.Subscription, e centrifuge.SubscribeSuccessEvent) {
	log.Printf("Subscribed on channel %s, resubscribed: %v, recovered: %v", sub.Channel(), e.Resubscribed, e.Recovered)
}

func (h *eventHandler) OnSubscribeError(sub *centrifuge.Subscription, e centrifuge.SubscribeErrorEvent) {
	log.Printf("Subscribed on channel %s failed, error: %s", sub.Channel(), e.Error)
}

func (h *eventHandler) OnUnsubscribe(sub *centrifuge.Subscription, _ centrifuge.UnsubscribeEvent) {
	log.Printf("Unsubscribed from channel %s", sub.Channel())
}

func New() *Client {

	handler := &eventHandler{}

	wsURL := "ws://centrifugo:8000/connection/websocket"
	c := centrifuge.New(wsURL, centrifuge.DefaultConfig())

	c.SetToken(connToken("micro-dadata", 0))

	c.OnConnect(handler)
	c.OnDisconnect(handler)
	c.OnMessage(handler)
	c.OnError(handler)

	c.OnServerPublish(handler)
	c.OnServerSubscribe(handler)
	c.OnServerUnsubscribe(handler)
	c.OnServerJoin(handler)
	c.OnServerLeave(handler)

	client := &Client{
		client:  c,
		handler: handler,
	}

	return client
}

func (c *Client) ListenChannel(microserviceName string, f func(*ResponseWriter)) {

	sub, err := c.client.NewSubscription("micro:" + microserviceName)
	if err != nil {
		log.Fatalln(err)
	}

	handler := &OnPublishHandler{
		client:  c.client,
		handler: f,
	}

	sub.OnPublish(handler)

	sub.OnJoin(c.handler)
	sub.OnLeave(c.handler)
	sub.OnSubscribeSuccess(c.handler)
	sub.OnSubscribeError(c.handler)
	sub.OnUnsubscribe(c.handler)

	err = sub.Subscribe()
	if err != nil {
		log.Fatalln(err)
	}

}

func (c *Client) Run() {

	err := c.client.Connect()
	if err != nil {
		log.Fatalln(err)
	}

	select {}

}
