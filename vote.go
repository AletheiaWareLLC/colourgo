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
	"github.com/golang/protobuf/proto"
)

func UnmarshalVote(data []byte) (*Vote, error) {
	vote := &Vote{}
	if err := proto.Unmarshal(data, vote); err != nil {
		return nil, err
	}
	return vote, nil
}

func GetVotes(votes *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, callback func(*bcgo.BlockEntry, *Vote) error) error {
	return bcgo.Iterate(votes.Name, votes.Head, nil, cache, network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			record := entry.Record
			v, err := UnmarshalVote(record.Payload)
			if err != nil {
				return err
			}
			callback(entry, v)
		}
		return nil
	})
}

func GetVotedColour(votes *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, x, y, z uint32) (*Colour, error) {
	var colours map[*Colour]int
	if err := bcgo.Iterate(votes.Name, votes.Head, nil, cache, network, func(hash []byte, block *bcgo.Block) error {
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

func CreateVote(w, x, y, z, red, green, blue, alpha uint32) *Vote {
	return &Vote{
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
	}
}

func CreateVoteRecord(alias string, key *rsa.PrivateKey, vote *Vote) (*bcgo.Record, error) {
	data, err := proto.Marshal(vote)
	if err != nil {
		return nil, err
	}
	return CreateRecord(alias, key, data)
}
