package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	database "github.com/codephobia/twitch-eos-thanks/app/database"
)

// find a twitch subscriber by id
func (t *Twitch) findSubscriberById(id string) (*database.Subscriber, error) {
	for _, v := range t.Subscribers {
		if v.SubscriberID == id {
			return v, nil
		}
	}

	return nil, fmt.Errorf("unable to find subscriber: %s", id)
}

func (t *Twitch) getSubscribers() error {
	log.Printf("[INFO] getSubscribers: checking api for subscribers")

	// get latest cached follow time
	latestSubscriberTime, err := t.getLatestSubscriberTime()
	if err != nil {
		return fmt.Errorf("latest subscriber time: %s", err)
	}

	i := 0
	loop := true

	for loop {
		offset := i * TWITCH_API_SUBSCRIBER_LIMIT

		// build out url
		u := []string{
			"http://",
			t.config.CodephobiaApiHost,
			":",
			t.config.CodephobiaApiPort,
			"/subscribers?channelID=",
			t.config.TwitchChannelID,
			"&latest=",
			strconv.FormatInt(latestSubscriberTime.UnixNano(), 10),
			"&limit=",
			strconv.Itoa(TWITCH_API_SUBSCRIBER_LIMIT),
		}

		// set offset if not on first page
		if i > 0 {
			u = append(u, strings.Join([]string{"&offset=", strconv.Itoa(offset)}, ""))
		}
		url := strings.Join(u, "")

		// get subscribers from server api
		body, err := t.getApiResponse(url)
		if err != nil {
			return err
		}

		// decode body
		subscriberResp := &SubscribersResp{}
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(subscriberResp); err != nil {
			return fmt.Errorf("body decode: %s", err)
		}

		// update subscribers
		t.Subscribers = subscriberResp.Data

		// check if we need to keep looping
		cnt := len(subscriberResp.Data)
		if cnt < TWITCH_API_SUBSCRIBER_LIMIT {
			// stop loop
			loop = false
		}

		// increment loop
		i++

		// sleep so we don't hammer api
		time.Sleep(TWITCH_API_DELAY)
	}

	log.Printf("[INFO] getSubscribers: found [%d] new subscribers", len(t.Subscribers))

	return nil
}

func (t *Twitch) getLatestSubscriberTime() (time.Time, error) {
	lt := time.Unix(0, 0)

	// get subscriber count
	count, err := t.database.Count(TWITCH_SUBSCRIBER_DB_BUCKET)
	if err != nil {
		return lt, fmt.Errorf("count: ", err)
	}

	// if we have cached subscribers, get latest subscribe time
	if count > 0 {
		// get all subscribers from database
		err, dbSubscribers := t.database.GetAll(TWITCH_SUBSCRIBER_DB_BUCKET)
		if err != nil {
			return lt, fmt.Errorf("get subscribers: ", err)
		}

		// loop through subscribers
		for _, dbSubscriber := range dbSubscribers {
			// unmarshal subscriber
			var subscriber database.Subscriber
			if err := json.Unmarshal(dbSubscriber, &subscriber); err != nil {
				return lt, fmt.Errorf("unmarshal subscriber: ", err)
			}

			// check if subscriber date is more recent
			if subscriber.Timestamp.After(lt) {
				lt = subscriber.Timestamp
			}
		}
	}

	return lt, nil
}

// the the subscribers to the database
func (t *Twitch) saveSubscribers() error {
	// check if we found subscribers
	if len(t.Subscribers) == 0 {
		return nil
	}

	for _, subscriber := range t.Subscribers {
		// put the subscriber data
		if err := t.database.Put(TWITCH_SUBSCRIBER_DB_BUCKET, subscriber.SubscriberID, *subscriber); err != nil {
			return fmt.Errorf("saving subscriber [%s]: %+v", err)
		}
	}

	// reset the subscribers
	t.Subscribers = make([]*database.Subscriber, 0)

	return nil
}
