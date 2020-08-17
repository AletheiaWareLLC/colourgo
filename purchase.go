/*
 * Copyright 2020 Aletheia Ware LLC
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
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/cryptogo"
	"github.com/golang/protobuf/proto"
)

func UnmarshalPurchase(data []byte) (*Purchase, error) {
	purchase := &Purchase{}
	if err := proto.Unmarshal(data, purchase); err != nil {
		return nil, err
	}
	return purchase, nil
}

func GetPurchases(purchases *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, callback func(*bcgo.BlockEntry, *Purchase) error) error {
	return bcgo.Iterate(purchases.Name, purchases.Head, nil, cache, network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			record := entry.Record
			p, err := UnmarshalPurchase(record.Payload)
			if err != nil {
				return err
			}
			callback(entry, p)
		}
		return nil
	})
}

func GetPurchasedColour(purchases *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, x, y, z uint32) (*Colour, error) {
	var colours map[*Colour]uint32
	var purchasedColour *Colour
	if err := bcgo.Iterate(purchases.Name, purchases.Head, nil, cache, network, func(hash []byte, block *bcgo.Block) error {
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

func CreatePurchase(w, x, y, z, red, green, blue, alpha, price, tax uint32) *Purchase {
	return &Purchase{
		Colour: &Colour{
			Red:   red,
			Green: green,
			Blue:  blue,
			Alpha: alpha,
		},
		Location: &Location{
			W: w,
			X: x,
			Y: y,
			Z: z,
		},
		Price: price,
		Tax:   tax,
	}
}

func CreatePurchaseRecord(alias string, key *rsa.PrivateKey, purchase *Purchase) (*bcgo.Record, error) {
	data, err := proto.Marshal(purchase)
	if err != nil {
		return nil, err
	}
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
