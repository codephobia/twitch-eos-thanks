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

func (t *Twitch) getBits() error {
	log.Printf("[INFO] getBits: checking api for bits")

	// get latest cached follow time
	latestBitsTime, err := t.getLatestBitsTime()
	if err != nil {
		return fmt.Errorf("latest bits time: %s", err)
	}

	i := 0
	loop := true

	for loop {
		offset := i * TWITCH_API_BITS_LIMIT

		// build out url
		u := []string{
			"http://",
			t.config.CodephobiaApiHost,
			":",
			t.config.CodephobiaApiPort,
			"/bits?channelID=",
			t.config.TwitchChannelID,
			"&latest=",
			strconv.FormatInt(latestBitsTime.UnixNano(), 10),
			"&limit=",
			strconv.Itoa(TWITCH_API_BITS_LIMIT),
		}

		// set offset if not on first page
		if i > 0 {
			u = append(u, strings.Join([]string{"&offset=", strconv.Itoa(offset)}, ""))
		}
		url := strings.Join(u, "")

		// get bits from server api
		body, err := t.getApiResponse(url)
		if err != nil {
			return err
		}

		// decode body
		bitsResp := &BitsResp{}
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(bitsResp); err != nil {
			return fmt.Errorf("body decode: %s", err)
		}

		// update bits
		t.Bits = append(t.Bits, bitsResp.Data...)

		// check if we need to keep looping
		cnt := len(bitsResp.Data)
		if cnt < TWITCH_API_BITS_LIMIT {
			// stop loop
			loop = false
		}

		// increment loop
		i++

		// sleep so we don't hammer api
		time.Sleep(TWITCH_API_DELAY)
	}

	log.Printf("[INFO] getBits: found [%d] new bits", len(t.Bits))

	return nil
}

func (t *Twitch) getLatestBitsTime() (time.Time, error) {
	lt := time.Unix(0, 0)

	// get bits count
	count, err := t.database.Count(TWITCH_BIT_DB_BUCKET)
	if err != nil {
		return lt, fmt.Errorf("count: %s", err)
	}

	// if we have cached bits, get latest bit time
	if count > 0 {
		// get all bits from database
		err, dbBits := t.database.GetAll(TWITCH_BIT_DB_BUCKET)
		if err != nil {
			return lt, fmt.Errorf("get bits: %s", err)
		}

		// loop through bits
		for _, dbBit := range dbBits {
			// unmarshal bit
			var bit database.Bit
			if err := json.Unmarshal(dbBit, &bit); err != nil {
				return lt, fmt.Errorf("unmarshal bit event: %s", err)
			}

			// check if bit date is more recent
			if bit.Time.After(lt) {
				lt = bit.Time
			}
		}
	}

	return lt, nil
}

// save the bits to the database
func (t *Twitch) saveBits() error {
	// check if we found bits
	if len(t.Bits) == 0 {
		return nil
	}

	for _, bit := range t.Bits {
		// put the bit data
		if err := t.database.Put(TWITCH_BIT_DB_BUCKET, bit.ID, *bit); err != nil {
			return fmt.Errorf("saving bit [%s]: %s", bit.ID, err)
		}
	}

	// reset the bits
	t.Bits = make([]*database.Bit, 0)

	return nil
}
