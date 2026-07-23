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

package converter

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/superdurable/iwf/gen/iwfpb"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/proto"
)

type plainNonProtoStruct struct {
	FlowType string `json:"flowType"`
}

func TestTemporalProtoPayloadIsBinaryProtobuf(t *testing.T) {
	dc := NewTemporalDataConverter()

	in := &iwfpb.InterpreterWorkflowInput{
		FlowType:     "order",
		WorkerTarget: "127.0.0.1:9000",
		StepInput:    &iwfpb.Value{Kind: &iwfpb.Value_StringValue{StringValue: "hi"}},
		InitAttributes: []*iwfpb.AttributeWrite{
			{Key: "k", Value: &iwfpb.Value{Kind: &iwfpb.Value_IntValue{IntValue: 7}}},
		},
	}

	payload, err := dc.ToPayload(in)
	require.NoError(t, err)
	require.Equal(t, converter.MetadataEncodingProto, string(payload.Metadata[converter.MetadataEncoding]))

	out := &iwfpb.InterpreterWorkflowInput{}
	require.NoError(t, dc.FromPayload(payload, out))
	require.True(t, proto.Equal(in, out))
}

func TestTemporalNilAndBytesRoundTrip(t *testing.T) {
	dc := NewTemporalDataConverter()

	nilPayload, err := dc.ToPayload(nil)
	require.NoError(t, err)
	var nilOut interface{}
	require.NoError(t, dc.FromPayload(nilPayload, &nilOut))
	require.Nil(t, nilOut)

	raw := []byte{1, 2, 3, 4}
	rawPayload, err := dc.ToPayload(raw)
	require.NoError(t, err)
	var rawOut []byte
	require.NoError(t, dc.FromPayload(rawPayload, &rawOut))
	require.Equal(t, raw, rawOut)
}

func TestTemporalMapOneofRoundTrip(t *testing.T) {
	dc := NewTemporalDataConverter()
	in := &iwfpb.PrepareRpcQueryResponse{
		RunId:        "run-1",
		FlowType:     "ft",
		WorkerTarget: "host:1",
		ChannelInfos: map[string]*iwfpb.ChannelInfo{
			"ch": {Size: 2},
		},
		Attributes: []*iwfpb.KV{
			{Key: "a", Value: &iwfpb.Value{Kind: &iwfpb.Value_BoolValue{BoolValue: true}}},
		},
	}
	payload, err := dc.ToPayload(in)
	require.NoError(t, err)
	require.Equal(t, converter.MetadataEncodingProto, string(payload.Metadata[converter.MetadataEncoding]))

	out := &iwfpb.PrepareRpcQueryResponse{}
	require.NoError(t, dc.FromPayload(payload, out))
	require.True(t, proto.Equal(in, out))
}

func TestTemporalJSONEscapeHatchForNonProto(t *testing.T) {
	dc := NewTemporalDataConverter()
	in := plainNonProtoStruct{FlowType: "should-be-json"}
	payload, err := dc.ToPayload(in)
	require.NoError(t, err)
	require.Equal(t, converter.MetadataEncodingJSON, string(payload.Metadata[converter.MetadataEncoding]))

	var out plainNonProtoStruct
	require.NoError(t, dc.FromPayload(payload, &out))
	require.Equal(t, in, out)
}

func TestMarshalDeterministicStableForMaps(t *testing.T) {
	msg := &iwfpb.ContinueAsNewDump{
		ChannelReceived: map[string]*iwfpb.ChannelValues{
			"b": {Values: []*iwfpb.Value{{Kind: &iwfpb.Value_StringValue{StringValue: "2"}}}},
			"a": {Values: []*iwfpb.Value{{Kind: &iwfpb.Value_StringValue{StringValue: "1"}}}},
		},
	}
	first, err := MarshalDeterministic(msg)
	require.NoError(t, err)
	second, err := MarshalDeterministic(msg)
	require.NoError(t, err)
	require.Equal(t, first, second)
}
