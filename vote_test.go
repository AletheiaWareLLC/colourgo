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

package colourgo_test

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/colourgo"
	"github.com/AletheiaWareLLC/testinggo"
	"testing"
	"time"
)

func awaitTrigger(t *testing.T, triggers chan bool) {
	select {
	case <-triggers:
	// Pass
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for trigger")
	}
}

func TestVoteModel_Trigger(t *testing.T) {
	cache := bcgo.NewMemoryCache(1)
	node := &bcgo.Node{
		Alias:    "TEST_ALIAS",
		Key:      nil,
		Cache:    cache,
		Network:  nil,
		Channels: make(map[string]*bcgo.Channel),
	}
	channel := &bcgo.Channel{
		Name: "TEST_CHANNEL",
	}
	canvas := &colourgo.Canvas{
		Name: "TEST_CANVAS",
	}
	id := "TEST_ID"
	triggers := make(chan bool, 1)
	model := colourgo.NewVoteModel(node, nil, id, canvas, channel, func() {
		triggers <- true
	})
	awaitTrigger(t, triggers)
	model.Trigger()
	awaitTrigger(t, triggers)
}

func TestVoteModel_Write(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Error("Could not generate key:", err)
	}
	cache := bcgo.NewMemoryCache(1)
	node := &bcgo.Node{
		Alias:    "TEST_ALIAS",
		Key:      key,
		Cache:    cache,
		Network:  nil,
		Channels: make(map[string]*bcgo.Channel),
	}
	channel := &bcgo.Channel{
		Name: "TEST_CHANNEL",
	}
	canvas := &colourgo.Canvas{
		Name: "TEST_CANVAS",
	}
	id := "TEST_ID"
	model := colourgo.NewVoteModel(node, nil, id, canvas, channel, nil)
	l := &colourgo.Location{
		X: 1,
		Y: 2,
		Z: 3,
	}
	c := &colourgo.Colour{
		Red:   0,
		Green: 1,
		Blue:  2,
		Alpha: 3,
	}
	testinggo.AssertNoError(t, model.Write(l, c))
	entries, err := cache.GetBlockEntries(channel.Name, 0)
	testinggo.AssertNoError(t, err)
	if len(entries) != 1 {
		t.Fatalf("Incorrect entries; expected 1, got '%d'", len(entries))
	}
	entry := entries[0]
	record := entry.Record
	vote, err := colourgo.UnmarshalVote(record.Payload)
	testinggo.AssertNoError(t, err)
	testinggo.AssertProtobufEqual(t, l, vote.Location)
	testinggo.AssertProtobufEqual(t, c, vote.Colour)
}

func TestFreeForAllModel_Draw(t *testing.T) {
}
