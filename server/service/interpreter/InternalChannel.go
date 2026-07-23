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
	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service/common/ptr"
)

type InternalChannel struct {
	// key is channel name
	receivedData map[string][]*iwfidl.EncodedObject
}

func NewInternalChannel() *InternalChannel {
	return &InternalChannel{
		receivedData: map[string][]*iwfidl.EncodedObject{},
	}
}

func RebuildInternalChannel(refill map[string][]*iwfidl.EncodedObject) *InternalChannel {
	return &InternalChannel{
		receivedData: refill,
	}
}

func (i *InternalChannel) GetAllReceived() map[string][]*iwfidl.EncodedObject {
	return i.receivedData
}

func (i *InternalChannel) GetInfos() map[string]iwfidl.ChannelInfo {
	infos := make(map[string]iwfidl.ChannelInfo, len(i.receivedData))
	for name, l := range i.receivedData {
		infos[name] = iwfidl.ChannelInfo{
			Size: ptr.Any(int32(len(l))),
		}
	}
	return infos
}

func (i *InternalChannel) HasData(channelName string) bool {
	l := i.receivedData[channelName]
	return len(l) > 0
}

func (i *InternalChannel) ProcessPublishing(publishes []iwfidl.InterStateChannelPublishing) {
	for _, pub := range publishes {
		i.receive(pub.ChannelName, pub.Value)
	}
}

func (i *InternalChannel) receive(channelName string, data *iwfidl.EncodedObject) {
	l := i.receivedData[channelName]
	l = append(l, data)
	i.receivedData[channelName] = l
}

func (i *InternalChannel) Retrieve(channelName string) *iwfidl.EncodedObject {
	l := i.receivedData[channelName]
	if len(l) <= 0 {
		panic("critical bug, this shouldn't happen")
	}
	data := l[0]
	l = l[1:]
	if len(l) == 0 {
		delete(i.receivedData, channelName)
	} else {
		i.receivedData[channelName] = l
	}

	return data
}
