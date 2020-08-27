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
	"github.com/AletheiaWareLLC/cryptogo"
	"time"
)

const (
	COLOUR = "Colour"

	COLOUR_THRESHOLD = bcgo.THRESHOLD_G

	COLOUR_HOST            = "colour.aletheiaware.com"
	COLOUR_HOST_TEST       = "test-colour.aletheiaware.com"
	COLOUR_PREFIX          = "Colour-"
	COLOUR_PREFIX_CANVAS   = "Colour-Canvas-"   // Append Year
	COLOUR_PREFIX_PURCHASE = "Colour-Purchase-" // Append Canvas ID
	COLOUR_PREFIX_VOTE     = "Colour-Vote-"     // Append Canvas ID
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

func GetPurchaseChannelName(id string) string {
	return COLOUR_PREFIX_PURCHASE + id
}

func GetVoteChannelName(id string) string {
	return COLOUR_PREFIX_VOTE + id
}

func OpenCanvasChannel() *bcgo.Channel {
	return bcgo.OpenPoWChannel(GetCanvasChannelName(), COLOUR_THRESHOLD)
}

func OpenPurchaseChannel(id string) *bcgo.Channel {
	return bcgo.OpenPoWChannel(GetPurchaseChannelName(id), COLOUR_THRESHOLD)
}

func OpenVoteChannel(id string) *bcgo.Channel {
	return bcgo.OpenPoWChannel(GetVoteChannelName(id), COLOUR_THRESHOLD)
}

func CreateRecord(alias string, key *rsa.PrivateKey, data []byte) (*bcgo.Record, error) {
	signature, err := cryptogo.CreateSignature(key, cryptogo.Hash(data), cryptogo.SignatureAlgorithm_SHA512WITHRSA_PSS)
	if err != nil {
		return nil, err
	}
	return &bcgo.Record{
		Timestamp:           bcgo.Timestamp(),
		Creator:             alias,
		Payload:             data,
		EncryptionAlgorithm: cryptogo.EncryptionAlgorithm_UNKNOWN_ENCRYPTION,
		Signature:           signature,
		SignatureAlgorithm:  cryptogo.SignatureAlgorithm_SHA512WITHRSA_PSS,
	}, nil
}
