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
	"fmt"
	"github.com/AletheiaWareLLC/bcgo"
	"sync"
)

type Model interface {
	Bind()
	Draw(func(*Location, *Colour))
	Mine() error
	Refresh() error
	Write(*Location, *Colour) error
}

func GetModel(node *bcgo.Node, listener bcgo.MiningListener, id string, canvas *Canvas, callback func()) (Model, error) {
	switch canvas.Mode {
	case Mode_FREE_FOR_ALL:
		name := GetVoteChannelName(id)
		channel := node.GetOrOpenChannel(name, func() *bcgo.Channel {
			return OpenVoteChannel(id)
		})
		return NewFreeForAllModel(node, listener, id, canvas, channel, callback), nil
		/* TODO
		   case Mode_DEMOCRACY:
		       name := GetVoteChannelName(id)
		       channel := m.Node.GetOrOpenChannel(name, func() *bcgo.Channel {
		           return OpenVoteChannel(id)
		       })
		       return NewDemocracyModel(node, listener, id, canvas, channel, callback), nil
		   case Mode_RADICAL_DEMOCRACY:
		       name := GetVoteChannelName(id)
		       channel := m.Node.GetOrOpenChannel(name, func() *bcgo.Channel {
		           return OpenVoteChannel(id)
		       })
		       return NewRadicalDemocracyModel(node, listener, id, canvas, channel, callback), nil
		   case Mode_MARKET:
		       name := GetPurchaseChannelName(id)
		       channel := m.Node.GetOrOpenChannel(name, func() *bcgo.Channel {
		           return OpenPurchaseChannel(id)
		       })
		       return NewMarketModel(node, listener, id, canvas, channel, callback), nil
		   case Mode_RADICAL_MARKET:
		       name := GetPurchaseChannelName(id)
		       channel := m.Node.GetOrOpenChannel(name, func() *bcgo.Channel {
		           return OpenPurchaseChannel(id)
		       })
		       return NewRadicalMarketModel(node, listener, id, canvas, channel, callback), nil
		*/
	case Mode_UNKNOWN_MODE:
		fallthrough
	default:
		return nil, fmt.Errorf("Unrecognized Canvas Mode: %s", canvas.Mode.String())
	}
}

type BaseModel struct {
	sync.Mutex
	Node     *bcgo.Node
	Listener bcgo.MiningListener
	ID       string
	Canvas   *Canvas
	Channel  *bcgo.Channel
	OnUpdate func()
	Entries  map[string]*bcgo.BlockEntry
	Order    []string
}

func NewBaseModel(node *bcgo.Node, listener bcgo.MiningListener, id string, canvas *Canvas, channel *bcgo.Channel, callback func()) *BaseModel {
	m := &BaseModel{
		Node:     node,
		Listener: listener,
		ID:       id,
		Canvas:   canvas,
		Channel:  channel,
		OnUpdate: callback,
		Entries:  make(map[string]*bcgo.BlockEntry),
	}
	go m.Refresh()
	return m
}

func (m *BaseModel) Bind() {
	// Do nothing
}

func (m *BaseModel) Draw(func(*Location, *Colour)) {
	// Do nothing
}

func (m *BaseModel) Write(*Location, *Colour) error {
	// Do nothing
	return nil
}

func (m *BaseModel) Refresh() error {
	return m.Channel.Refresh(m.Node.Cache, m.Node.Network)
}

func (m *BaseModel) Mine() error {
	// Mine Channel
	if _, _, err := m.Node.Mine(m.Channel, COLOUR_THRESHOLD, m.Listener); err != nil {
		return err
	}

	if m.Node.Network != nil {
		// Push Update to Peers
		if err := m.Channel.Push(m.Node.Cache, m.Node.Network); err != nil {
			return err
		}
	}
	return nil
}
