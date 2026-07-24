// Copyright (c) 2022-2026 Super Durable, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package interpreter

import (
	"fmt"

	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service/interpreter/channel"
)

// ChannelStore holds FIFO messages by channel.
type ChannelStore struct {
	receivedData map[string][]*iwfpb.Value
}

func NewChannelStore() *ChannelStore {
	return &ChannelStore{receivedData: map[string][]*iwfpb.Value{}}
}

// RebuildChannelStore restores a snapshot.
func RebuildChannelStore(refill map[string]*iwfpb.ChannelValues) *ChannelStore {
	data := make(map[string][]*iwfpb.Value, len(refill))
	for name, channelValues := range refill {
		if len(channelValues.GetValues()) > 0 {
			data[name] = channelValues.GetValues()
		}
	}
	return &ChannelStore{receivedData: data}
}

// Publish appends messages.
func (s *ChannelStore) Publish(messages []*iwfpb.ChannelMessage) {
	for _, message := range messages {
		channelName := message.GetChannelName()
		s.receivedData[channelName] = append(s.receivedData[channelName], message.GetValue())
	}
}

// Availability returns an isolated count snapshot.
func (s *ChannelStore) Availability() channel.ChannelAvailability {
	availability := make(channel.ChannelAvailability, len(s.receivedData))
	for name, values := range s.receivedData {
		availability[name] = int32(len(values))
	}
	return availability
}

// HasAtLeast tests a channel size.
func (s *ChannelStore) HasAtLeast(channelName string, n int32) bool {
	return int32(len(s.receivedData[channelName])) >= n
}

// HasData reports whether a channel has messages.
func (s *ChannelStore) HasData(channelName string) bool {
	return len(s.receivedData[channelName]) > 0
}

// GetInfos returns channel sizes.
func (s *ChannelStore) GetInfos() map[string]*iwfpb.ChannelInfo {
	infos := make(map[string]*iwfpb.ChannelInfo, len(s.receivedData))
	for name, values := range s.receivedData {
		infos[name] = &iwfpb.ChannelInfo{Size: int32(len(values))}
	}
	return infos
}

// Dump returns a snapshot.
func (s *ChannelStore) Dump() map[string]*iwfpb.ChannelValues {
	snapshot := make(map[string]*iwfpb.ChannelValues, len(s.receivedData))
	for name, values := range s.receivedData {
		snapshot[name] = &iwfpb.ChannelValues{Values: values}
	}
	return snapshot
}

// CommitMatch consumes a plan and returns the consumed messages.
func (s *ChannelStore) CommitMatch(plan *channel.MatchPlan) map[int][]*iwfpb.Value {
	consumed := make(map[int][]*iwfpb.Value, len(plan.Consumes))
	for _, consumption := range plan.Consumes {
		values := s.receivedData[consumption.ChannelName]
		if int32(len(values)) < consumption.Count {
			panic(fmt.Sprintf(
				"channel %q holds %d messages but the match plan consumes %d; no yield may occur between plan and commit",
				consumption.ChannelName,
				len(values),
				consumption.Count,
			))
		}
		if consumption.Count > 0 {
			consumed[consumption.ChannelConditionIndex] = values[:consumption.Count:consumption.Count]
			values = values[consumption.Count:]
		} else {
			consumed[consumption.ChannelConditionIndex] = nil
		}
		if len(values) == 0 {
			delete(s.receivedData, consumption.ChannelName)
		} else {
			s.receivedData[consumption.ChannelName] = values
		}
	}
	return consumed
}
