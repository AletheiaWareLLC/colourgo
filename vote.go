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

func UnmarshalVote(data []byte) (*Vote, error) {
	vote := &Vote{}
	if err := proto.Unmarshal(data, vote); err != nil {
		return nil, err
	}
	return vote, nil
}

func GetVotes(cs bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string) ([]*Vote, error) {
	votes := make([]*Vote, 0)
	if err := bcgo.Iterate(cs.Name, cs.Head, nil, cache, network, func(hash []byte, block *bcgo.Block) error {
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
	if err := bcgo.Iterate(cs.Name, cs.Head, nil, cache, network, func(hash []byte, block *bcgo.Block) error {
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

func CreateVoteRecord(alias string, key *rsa.PrivateKey, w, x, y, z, red, green, blue, alpha uint32) (*bcgo.Record, error) {
	data, err := proto.Marshal(&Vote{
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
	})
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
