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
	return "https://" + GetColourHost()
}

func GetYear() string {
	return fmt.Sprintf("%d", time.Now().Year())
}

func OpenCanvasChannel() *bcgo.PoWChannel {
	return bcgo.OpenPoWChannel(COLOUR_PREFIX_CANVAS+GetYear(), bcgo.THRESHOLD_STANDARD)
}

func UnmarshalCanvas(data []byte) (*Canvas, error) {
	canvas := &Canvas{}
	if err := proto.Unmarshal(data, canvas); err != nil {
		return nil, err
	}
	return canvas, nil
}

func UnmarshalPurchase(data []byte) (*Purchase, error) {
	purchase := &Purchase{}
	if err := proto.Unmarshal(data, purchase); err != nil {
		return nil, err
	}
	return purchase, nil
}

func UnmarshalVote(data []byte) (*Vote, error) {
	vote := &Vote{}
	if err := proto.Unmarshal(data, vote); err != nil {
		return nil, err
	}
	return vote, nil
}

func GetCanvas(canvases bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string, key *rsa.PrivateKey, recordHash []byte, callback func(*bcgo.BlockEntry, []byte, *Canvas) error) error {
	return bcgo.Read(canvases.GetName(), canvases.GetHead(), nil, cache, network, alias, key, recordHash, func(entry *bcgo.BlockEntry, key, data []byte) error {
		canvas, err := UnmarshalCanvas(data)
		if err != nil {
			return err
		}
		return callback(entry, key, canvas)
	})
}

func GetVotes(cs bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string) ([]*Vote, error) {
	votes := make([]*Vote, 0)
	if err := bcgo.Iterate(cs.GetName(), cs.GetHead(), nil, cache, network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			record := entry.Record
			if record.Creator == alias {
				v, err := UnmarshalVote(record.Payload)
				if err != nil {
					return err
				}
				votes = append(votes, v)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return votes, nil
}

func GetVotedColour(cs bcgo.Channel, cache bcgo.Cache, network bcgo.Network, x, y, z uint32) (*Colour, error) {
	var colours map[*Colour]int
	if err := bcgo.Iterate(cs.GetName(), cs.GetHead(), nil, cache, network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			record := entry.Record
			v, err := UnmarshalVote(record.Payload)
			if err != nil {
				return err
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
		return nil
	}); err != nil {
		return nil, err
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

func GetPurchases(cs bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string) ([]*Purchase, error) {
	purchases := make([]*Purchase, 0)
	if err := bcgo.Iterate(cs.GetName(), cs.GetHead(), nil, cache, network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			record := entry.Record
			if record.Creator == alias {
				p, err := UnmarshalPurchase(record.Payload)
				if err != nil {
					return err
				}
				purchases = append(purchases, p)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return purchases, nil
}

func GetPurchasedColour(cs bcgo.Channel, cache bcgo.Cache, network bcgo.Network, x, y, z uint32) (*Colour, error) {
	var colours map[*Colour]uint32
	var purchasedColour *Colour
	if err := bcgo.Iterate(cs.GetName(), cs.GetHead(), nil, cache, network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			record := entry.Record
			p, err := UnmarshalPurchase(record.Payload)
			if err != nil {
				return err
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
		return nil
	}); err != nil {
		return nil, err
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
