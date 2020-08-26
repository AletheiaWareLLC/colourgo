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
	"github.com/AletheiaWareLLC/bcgo"
	"sync"
)

type Model interface {
	Draw(func(*Location, *Colour))
	Mine() error
	Refresh() error
	Write(*Location, *Colour) error
}

type BaseModel struct {
	sync.Mutex
	Node       *bcgo.Node
	Listener   bcgo.MiningListener
	ID         string
	Canvas     *Canvas
	Channel    *bcgo.Channel
	IsUpdating bool
	OnUpdate   func()
	Entries    map[string]*bcgo.BlockEntry
	Order      []string
}

func NewBaseModel(node *bcgo.Node, listener bcgo.MiningListener, id string, canvas *Canvas, channel *bcgo.Channel) *BaseModel {
	return &BaseModel{
		Node:     node,
		Listener: listener,
		ID:       id,
		Canvas:   canvas,
		Channel:  channel,
		Entries:  make(map[string]*bcgo.BlockEntry),
	}
}

func (m *BaseModel) Draw(func(*Location, *Colour)) {
	// Do nothing
}

func (m *BaseModel) Write(*Location, *Colour) error {
	// Do nothing
	return nil
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

func (m *BaseModel) Refresh() error {
	// Load Channel
	if err := m.Channel.LoadCachedHead(m.Node.Cache); err != nil {
		return err
	}
	if m.Node.Network != nil {
		// Pull Channel
		if err := m.Channel.Pull(m.Node.Cache, m.Node.Network); err != nil {
			return err
		}
	}
	return nil
}