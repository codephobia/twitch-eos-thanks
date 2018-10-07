package twitch

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/codephobia/twitch-eos-thanks/server/config"
	"github.com/codephobia/twitch-eos-thanks/server/database"
)

var (
	bearerPrefix   = "Bearer "
	pubsubURL      = "wss://pubsub-edge.twitch.tv"
	writeWait      = 1 * time.Second
	pingPeriod     = 5 * time.Minute
	pongWait       = 10 * time.Second
	maxMessageSize = int64(512)

	backoffMin    = 1 * time.Second
	backoffMax    = 2 * time.Minute
	backoffFactor = float64(2)
	backoffJitter = true
)

// PUBSUB is a pub sub manager for twitch.
type PUBSUB struct {
	config   *config.Config
	database *database.Database
	twitch   *Twitch

	ctx       context.Context
	ctxCancel context.CancelFunc

	client *http.Client
	conn   *websocket.Conn
	Send   chan []byte

	pongTimer *time.Timer
	pongDone  chan bool

	Backoff *Backoff
}

// NewPUBSUB returns a new pub sub.
func NewPUBSUB(c *config.Config, db *database.Database, t *Twitch) *PUBSUB {
	ctx, cancel := context.WithCancel(context.Background())

	return &PUBSUB{
		config:   c,
		database: db,
		twitch:   t,

		ctx:       ctx,
		ctxCancel: cancel,

		client:   &http.Client{},
		Send:     make(chan []byte, 256),
		pongDone: make(chan bool),

		Backoff: &Backoff{
			Min:    backoffMin,
			Max:    backoffMax,
			Factor: backoffFactor,
			Jitter: backoffJitter,
		},
	}
}

// Init initializes the pub sub listener.
func (p *PUBSUB) Init() error {
	log.Printf("[INFO] pubsub: initializing")

	// connect to twitch pubsub
	if err := p.connect(); err != nil {
		return fmt.Errorf("connect: %s", err)
	}

	// enable read / write
	go p.ReadPump()
	go p.WritePump()

	if err := p.listenRequest(); err != nil {
		return fmt.Errorf("listen request: %s", err)
	}

	// return with no error
	return nil
}

// connect to twitch pub sub
func (p *PUBSUB) connect() error {
	log.Printf("[INFO] pubsub: connecting")

	// create auth headers
	headers := http.Header{"Authorization": {bearerPrefix + p.config.TwitchOAuthToken}}

	// dial connection
	conn, _, err := websocket.DefaultDialer.DialContext(p.ctx, pubsubURL, headers)
	if err != nil {
		return fmt.Errorf("unable to dial connection: %s", err)
	}

	// save connection
	p.conn = conn

	return nil
}

// sends the listen request to twitch
func (p *PUBSUB) listenRequest() error {
	log.Printf("[INFO] pubsub: sending listen request")

	// create subs listen request
	subsTopics := []string{
		strings.Join([]string{PUBSUBTopicSubscription.String(), p.config.TwitchChannelID}, "."),
	}
	subsReq := NewPUBSUBRequest(PUBSUBTypeListen.String(), subsTopics, p.config.TwitchChannelOAuthToken)

	// convert request to bytes
	reqBytes, err := subsReq.ToBytes()
	if err != nil {
		return err
	}

	// send request
	p.Send <- reqBytes

	return nil
}

// ReadPump reads incoming messages on the websocket connection.
func (p *PUBSUB) ReadPump() {
	defer func() {
		log.Printf("[INFO] pubsub: closing read")
		p.conn.Close()
	}()

	p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadDeadline(time.Now().Add(pongWait + pingPeriod))
	p.conn.SetPongHandler(func(string) error { p.conn.SetReadDeadline(time.Now().Add(pongWait + pingPeriod)); return nil })

	for {
		_, message, err := p.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("[ERROR] unexpected close error: %s", err)
				return
			}
			break
		}

		// convert bytes to message
		msg, err := NewPUBSUBMessage(message)
		if err != nil {
			log.Printf("[ERROR] pub sub message: %s", err)
			continue
		}

		log.Printf("[INFO] message: %+v", msg)

		// handle message
		p.handleWSMessage(msg)
	}
}

// WritePump writes outgoing messages on the websocket connection.
func (p *PUBSUB) WritePump() {
	// TODO: add jitter to ping timer
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Printf("[INFO] pubsub: closing write")
		ticker.Stop()
		p.conn.Close()
	}()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))

			// send ws ping
			if err := p.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("[ERROR] sending ping: %s", err)
				return
			}

			// ping twitch
			p.pingTwitch()
		case message, ok := <-p.Send:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// close connection
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := p.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

// handle incoming websocket messages
func (p *PUBSUB) handleWSMessage(msg *PUBSUBMessage) {
	switch msg.Type {
	case PUBSUBTypeResponse:
		// return out if empty error
		if len(msg.Error) == 0 {
			return
		}

		// handle error
		p.handleResponseError(msg.Error)
	case PUBSUBTypeMessage:
		log.Printf("[INFO] pubsub: message: %+v", msg.Data)
		p.handleMessage(msg)
	case PUBSUBTypePong:
		log.Printf("[INFO] pubsub: PONG")
		p.pongDone <- true
	case PUBSUBTypeReconnect:
		log.Printf("[INFO] pubsub: reconnect alert received")
		p.reconnect()
	}
}

func (p *PUBSUB) handleResponseError(err string) {
	switch PUBSUBMessageError(err) {
	// bad auth token
	case errBadAuth:
		log.Printf("[ERROR] pubsub: response: bad auth: %s", err)

		// refresh oauth token
		if err := p.refreshToken(); err != nil {
			log.Printf("[ERROR] pubsub: %s", err)
			return
		}

		// reconnect to twitch websocket
		p.reconnect()
	case errBadMessage:
		log.Printf("[ERROR] pubsub: response: bad message: %s", err)
	case errBadTopic:
		log.Printf("[ERROR] pubsub: response: bad topic: %s", err)
	case errServer2:
		fallthrough
	case errServer:
		log.Printf("[ERROR] pubsub: response: server: %s", err)
	}
}

// handle a websocket message of type MESSAGE
func (p *PUBSUB) handleMessage(msg *PUBSUBMessage) {
	// split topic from channel id
	msgTopic := strings.Split(msg.Data.Topic, ".")

	// validate topic split length
	if len(msgTopic) != 2 {
		log.Printf("invalid message topic: %+v", msgTopic)
		return
	}

	topic := msgTopic[0]
	channelID := msgTopic[1]

	switch PUBSUBTopic(topic) {
	case PUBSUBTopicSubscription:
		subscription, err := NewPUBSUBSubscriptionMessage(msg.Data.Message)
		if err != nil {
			log.Printf("[ERROR] sub message: %s", err)
			return
		}

		// generate emotes for database
		messageEmotes := make([]*database.SubMessageEmote, 0)

		// loop through sub emotes
		for _, emote := range subscription.SubMessage.Emotes {
			messageEmotes = append(messageEmotes, &database.SubMessageEmote{
				Start: emote.Start,
				End:   emote.End,
				ID:    emote.ID,
			})
		}

		// convert timestamp
		timestamp, err := time.Parse(time.RFC3339, subscription.Time)
		if err != nil {
			log.Printf("unable to convert sub timestamp: %s", err)
		}

		// add the subscriber to the database
		if err := p.database.AddSubscriber(&database.Subscriber{
			ChannelID:    channelID,
			SubscriberID: subscription.UserID,
			Timestamp:    timestamp,

			DisplayName: subscription.DisplayName,
			SubPlan:     subscription.SubPlan,
			SubPlanName: subscription.SubPlanName,
			Months:      subscription.Months,
			Context:     subscription.Context,
			SubMessage: &database.SubMessage{
				Message: subscription.SubMessage.Message,
				Emotes:  messageEmotes,
			},
		}); err != nil {
			log.Printf("[ERROR] add sub: %s", err)
		}

		return
	case PUBSUBTopicBits:
		log.Printf("bits: %s", msg.Data.Message)
		return
	case PUBSUBTopicCommerce:
		log.Printf("commerce: %s", msg.Data.Message)
		return
	case PUBSUBTopicWhispers:
		log.Printf("whispers: %s", msg.Data.Message)
		return
	}
}

// sends a ping to twitch over the websocket
func (p *PUBSUB) pingTwitch() {
	// send twitch ping
	ping := &PUBSUBMessage{
		Type: PUBSUBTypePing,
	}

	// convert ping to bytes
	pingBytes, err := ping.ToBytes()
	if err != nil {
		log.Printf("[ERROR] unable to generate ping: %s", err)
		return
	}

	// send ping
	p.Send <- pingBytes

	// start pong check
	p.pongCheck()
}

// checks if we receive a pong within 10 seconds
// otherwise reconnect the websocket connection
func (p *PUBSUB) pongCheck() {
	// setup pong checker
	p.pongTimer = time.NewTimer(pongWait)

	go func() {
		select {
		case <-p.pongTimer.C:
			// pong timer lapsed so now we reconnect
			log.Printf("[INFO] PONG timeout: reconnecting")
			p.reconnect()
		case <-p.ctx.Done():
		case <-p.pongDone:
			// pong received so stop timer
			log.Printf("[INFO] PONG received in time")
			p.pongTimer.Stop()
		}
	}()
}

// reconnects an existing connection to twitch pub sub
func (p *PUBSUB) reconnect() {
	log.Printf("[INFO] pubsub: reconnect")

	// cancel current connection context
	p.ctxCancel()

	// update context
	p.ctx, p.ctxCancel = context.WithCancel(context.Background())

	// create backoff timer
	timer := time.NewTimer(p.Backoff.Duration())

	// wait for timer before attempting reconnect
	go func(timer *time.Timer) {
		select {
		case <-timer.C:
			p.attemptReconnect()
		}
	}(timer)
}

func (p *PUBSUB) attemptReconnect() {
	// connect to twitch
	if err := p.connect(); err != nil {
		log.Printf("[ERROR] connect: %s", err)
		p.reconnect()
		return
	}

	// reset backoff
	p.Backoff.Reset()

	// enable read / write
	go p.ReadPump()
	go p.WritePump()

	// send listen request
	if err := p.listenRequest(); err != nil {
		// TODO: gracefully fail, and re-attempt
		log.Printf("[ERROR] listen request: %s", err)
		return
	}
}
