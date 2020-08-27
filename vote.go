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
	"encoding/base64"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/golang/protobuf/proto"
	"log"
	"sort"
)

type VoteModel struct {
	BaseModel
	Votes map[string]*Vote
}

func NewVoteModel(node *bcgo.Node, listener bcgo.MiningListener, id string, canvas *Canvas, channel *bcgo.Channel, callback func()) *VoteModel {
	m := &VoteModel{
		BaseModel: *NewBaseModel(node, listener, id, canvas, channel, callback),
		Votes:     make(map[string]*Vote),
	}
	m.Channel.AddTrigger(m.Trigger)
	go m.Trigger()
	return m
}

func (m *VoteModel) Trigger() {
	log.Println("Trigger:", m.Channel.Name)
	m.Lock()
	if err := GetVotes(m.Channel, m.Node.Cache, m.Node.Network, func(entry *bcgo.BlockEntry, vote *Vote) error {
		id := base64.RawURLEncoding.EncodeToString(entry.RecordHash)
		log.Println("Got Vote:", id, entry.Record.Timestamp, vote)
		_, ok := m.Votes[id]
		if ok {
			log.Println("Vote already counted")
			return bcgo.StopIterationError{}
		} else {
			m.Votes[id] = vote
			m.Entries[id] = entry
			m.Order = append(m.Order, id)
		}
		return nil
	}); err != nil {
		switch err.(type) {
		case bcgo.StopIterationError:
			// Do nothing
		default:
			log.Println(err)
		}
	}
	sort.Slice(m.Order, func(i, j int) bool {
		return m.Entries[m.Order[i]].Record.Timestamp < m.Entries[m.Order[j]].Record.Timestamp
	})
	m.Unlock()
	go func() {
		if f := m.OnUpdate; f != nil {
			f()
		}
		if err := m.Mine(); err != nil {
			log.Println(err)
		}
	}()
}

func (m *VoteModel) Write(l *Location, c *Colour) error {
	record, err := CreateVoteRecord(m.Node.Alias, m.Node.Key, &Vote{
		Colour:   c,
		Location: l,
	})
	if err != nil {
		return err
	}
	_, err = bcgo.WriteRecord(m.Channel.Name, m.Node.Cache, record)
	if err != nil {
		return err
	}
	return nil
}

type FreeForAllModel struct {
	VoteModel
}

func NewFreeForAllModel(node *bcgo.Node, listener bcgo.MiningListener, id string, canvas *Canvas, channel *bcgo.Channel, callback func()) *FreeForAllModel {
	return &FreeForAllModel{
		VoteModel: *NewVoteModel(node, listener, id, canvas, channel, callback),
	}
}

func (m *FreeForAllModel) Draw(callback func(*Location, *Colour)) {
	log.Println("Drawing:", m.Order)
	for _, id := range m.Order {
		vote, ok := m.Votes[id]
		if ok {
			log.Println("Drawing Vote:", id, m.Entries[id].Record.Timestamp, vote)
			callback(vote.Location, vote.Colour)
		}
	}
}

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
