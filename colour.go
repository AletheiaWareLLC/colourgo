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

	COLOUR_HOST            = "colour.aletheiaware.com"
	COLOUR_HOST_TEST       = "test-colour.aletheiaware.com"
	COLOUR_WEBSITE         = "https://colour.aletheiaware.com"
	COLOUR_WEBSITE_TEST    = "https://test-colour.aletheiaware.com"
	COLOUR_PREFIX_CANVAS   = "Colour-Canvas-" // Append Year
	COLOUR_PREFIX_PURCHASE = "Colour-Purchase-"
	COLOUR_PREFIX_VOTE     = "Colour-Vote-"
)

func GetColourHost() string {
	if bcgo.IsDebug() {
		return COLOUR_HOST_TEST
	}
	return COLOUR_HOST
}

func GetColourWebsite() string {
	if bcgo.IsDebug() {
		return COLOUR_WEBSITE_TEST
	}
	return COLOUR_WEBSITE
}

func GetYear() string {
	return fmt.Sprintf("%d", time.Now().Year())
}

func OpenColourChannel() (*bcgo.Channel, error) {
	return bcgo.OpenChannel(COLOUR)
}

func UnmarshalCanvas(data []byte) (*Canvas, error) {
	canvas := &Canvas{}
	if err := proto.Unmarshal(data, canvas); err != nil {
		return nil, err
	}
	return canvas, nil
}

func GetCanvas(canvases *bcgo.Channel, alias string, key *rsa.PrivateKey, recordHash []byte, callback func(*bcgo.BlockEntry, []byte, *Canvas) error) error {
	return canvases.Read(alias, key, recordHash, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Canvas
		canvas, err := UnmarshalCanvas(data)
		if err != nil {
			return err
		}
		return callback(entry, key, canvas)
	})
}

func GetVotes(cs *bcgo.Channel, alias string) ([]*Vote, error) {
	votes := make([]*Vote, 0)
	b := cs.HeadBlock
	for b != nil {
		for _, e := range b.Entry {
			r := e.Record
			if r.Creator == alias {
				v := &Vote{}
				err := proto.Unmarshal(r.Payload, v)
				if err != nil {
					return nil, err
				}
				votes = append(votes, v)
			}
		}
		h := b.Previous
		if h != nil && len(h) > 0 {
			var err error
			b, err = bcgo.ReadBlockFile(cs.Cache, h)
			if err != nil {
				return nil, err
			}
		} else {
			b = nil
		}
	}
	return votes, nil
}

func GetVotedColour(cs *bcgo.Channel, x, y, z uint32) (*Colour, error) {
	var colours map[*Colour]int
	b := cs.HeadBlock
	for b != nil {
		for _, e := range b.Entry {
			r := e.Record
			v := &Vote{}
			err := proto.Unmarshal(r.Payload, v)
			if err != nil {
				return nil, err
			}
			l := v.Location
			if l.X == x && l.Y == y && l.Z == z {
				count := 1
				if val, ok := colours[v.Colour]; ok {
					count = count + val
				}
				colours[v.Colour] = count
			}
		}
		h := b.Previous
		if h != nil && len(h) > 0 {
			var err error
			b, err = bcgo.ReadBlockFile(cs.Cache, h)
			if err != nil {
				return nil, err
			}
		} else {
			b = nil
		}
	}
	var maxColour *Colour
	maxCount := 0
	for colour, count := range colours {
		if count > maxCount {
			maxColour = colour
			maxCount = count
		}
	}
	return maxColour, nil
}

func CreateVoteRecord(alias string, key *rsa.PrivateKey, x, y, z, red, green, blue uint32) (*bcgo.Record, error) {
	data, err := proto.Marshal(&Vote{
		Colour: &Colour{
			Red:   red,
			Green: green,
			Blue:  blue,
		},
		Location: &Location{
			X: x,
			Y: y,
			Z: z,
		},
	})
	if err != nil {
		return nil, err
	}

	signature, err := bcgo.CreateSignature(key, bcgo.Hash(data), bcgo.SignatureAlgorithm_SHA512WITHRSA_PSS)
	if err != nil {
		return nil, err
	}

	return &bcgo.Record{
		Timestamp:           uint64(time.Now().UnixNano()),
		Creator:             alias,
		Payload:             data,
		EncryptionAlgorithm: bcgo.EncryptionAlgorithm_UNKNOWN_ENCRYPTION,
		Signature:           signature,
		SignatureAlgorithm:  bcgo.SignatureAlgorithm_SHA512WITHRSA_PSS,
	}, nil
}

func GetPurchases(cs *bcgo.Channel, alias string) ([]*Purchase, error) {
	purchases := make([]*Purchase, 0)
	b := cs.HeadBlock
	for b != nil {
		for _, e := range b.Entry {
			r := e.Record
			if r.Creator == alias {
				p := &Purchase{}
				err := proto.Unmarshal(r.Payload, p)
				if err != nil {
					return nil, err
				}
				purchases = append(purchases, p)
			}
		}
		h := b.Previous
		if h != nil && len(h) > 0 {
			var err error
			b, err = bcgo.ReadBlockFile(cs.Cache, h)
			if err != nil {
				return nil, err
			}
		} else {
			b = nil
		}
	}
	return purchases, nil
}

func GetPurchasedColour(cs *bcgo.Channel, x, y, z uint32) (*Colour, error) {
	var colours map[*Colour]uint32
	var purchasedColour *Colour
	b := cs.HeadBlock
	// For Each Existing Block
	for b != nil {
		// For Each Existing Entry
		for _, e := range b.Entry {
			// Reach Purchase
			r := e.Record
			p := &Purchase{}
			err := proto.Unmarshal(r.Payload, p)
			if err != nil {
				return nil, err
			}
			l := p.Location
			// If Location Matches Query
			if l.X == x && l.Y == y && l.Z == z {
				val, ok := colours[p.Colour]
				if !ok {
					colours[p.Colour] = p.Price
				} else {
					// If Price is the Highest Bid
					if val < p.Price {
						// Bought
						purchasedColour = p.Colour
						colours[p.Colour] = p.Price
					}
				}
			}
		}
		h := b.Previous
		if h != nil && len(h) > 0 {
			var err error
			b, err = bcgo.ReadBlockFile(cs.Cache, h)
			if err != nil {
				return nil, err
			}
		} else {
			b = nil
		}
	}
	return purchasedColour, nil
}

func CreatePurchaseRecord(alias string, key *rsa.PrivateKey, x, y, z, red, green, blue, price, tax uint32) (*bcgo.Record, error) {
	data, err := proto.Marshal(&Purchase{
		Colour: &Colour{
			Red:   red,
			Green: green,
			Blue:  blue,
		},
		Location: &Location{
			X: x,
			Y: y,
			Z: z,
		},
		Price: price,
		Tax:   tax,
	})
	if err != nil {
		return nil, err
	}

	signature, err := bcgo.CreateSignature(key, bcgo.Hash(data), bcgo.SignatureAlgorithm_SHA512WITHRSA_PSS)
	if err != nil {
		return nil, err
	}

	return &bcgo.Record{
		Timestamp:           uint64(time.Now().UnixNano()),
		Creator:             alias,
		Payload:             data,
		EncryptionAlgorithm: bcgo.EncryptionAlgorithm_UNKNOWN_ENCRYPTION,
		Signature:           signature,
		SignatureAlgorithm:  bcgo.SignatureAlgorithm_SHA512WITHRSA_PSS,
	}, nil
}
