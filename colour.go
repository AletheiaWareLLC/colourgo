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
	"fmt"
	"github.com/AletheiaWareLLC/bcgo"
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

func GetCanvasChannelName() string {
	return COLOUR_PREFIX_CANVAS + GetYear()
}

func OpenCanvasChannel() *bcgo.Channel {
	return bcgo.OpenPoWChannel(GetCanvasChannelName(), COLOUR_THRESHOLD)
}

func OpenPurchaseChannel(id string) *bcgo.Channel {
	return bcgo.OpenPoWChannel(COLOUR_PREFIX_PURCHASE+id, COLOUR_THRESHOLD)
}

func OpenVoteChannel(id string) *bcgo.Channel {
	return bcgo.OpenPoWChannel(COLOUR_PREFIX_VOTE+id, COLOUR_THRESHOLD)
}
