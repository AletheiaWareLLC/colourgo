/*
 * Copyright 2019 Aletheia Ware LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package colourgo

import (
	"crypto/rsa"
	"fmt"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/golang/protobuf/proto"
	"time"
)

const (
	COLOUR = "Colour"

	COLOUR_THRESHOLD = bcgo.THRESHOLD_G

	COLOUR_HOST            = "colour.aletheiaware.com"
	COLOUR_HOST_TEST       = "test-colour.aletheiaware.com"
	COLOUR_PREFIX_CANVAS   = "Colour-Canvas-" // Append Year
	COLOUR_PREFIX_PURCHASE = "Colour-Purchase-"
	COLOUR_PREFIX_VOTE     = "Colour-Vote-"
)

func GetColourHost() string {
	if bcgo.IsLive() {
		return COLOUR_HOST
	}
	return COLOUR_HOST_TEST
}

func GetColourWebsite() string {
	return "https://" + GetColourHost()
}

func GetYear() string {
	return fmt.Sprintf("%d", time.Now().UTC().Year())
}

func OpenCanvasChannel() *bcgo.Channel {
	return bcgo.OpenPoWChannel(COLOUR_PREFIX_CANVAS+GetYear(), COLOUR_THRESHOLD)
}

func UnmarshalCanvas(data []byte) (*Canvas, error) {
	canvas := &Canvas{}
	if err := proto.Unmarshal(data, canvas); err != nil {
		return nil, err
	}
	return canvas, nil
}

func GetCanvas(canvases bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string, key *rsa.PrivateKey, recordHash []byte, callback func(*bcgo.BlockEntry, []byte, *Canvas) error) error {
	return bcgo.Read(canvases.Name, canvases.Head, nil, cache, network, alias, key, recordHash, func(entry *bcgo.BlockEntry, key, data []byte) error {
		canvas, err := UnmarshalCanvas(data)
		if err != nil {
			return err
		}
		return callback(entry, key, canvas)
	})
}
