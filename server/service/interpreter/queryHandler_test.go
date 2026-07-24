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
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
)

type s2QueryRegistrar struct {
	queryType string
	handler   interface{}
}

func (r *s2QueryRegistrar) SetQueryHandler(
	_ interfaces.UnifiedContext,
	queryType string,
	handler interface{},
) error {
	r.queryType = queryType
	r.handler = handler
	return nil
}

func TestSetGetAttributesQueryHandler(t *testing.T) {
	provider := &s2WorkflowProvider{}
	manager, err := NewPersistenceManager(provider, []*iwfpb.KV{
		stringKV("b", "two"),
		stringKV("a", "one"),
	})
	require.NoError(t, err)

	registrar := &s2QueryRegistrar{}
	require.NoError(t, SetGetAttributesQueryHandler(nil, registrar, manager))
	require.Equal(t, service.GetAttributesWorkflowQueryType, registrar.queryType)

	handler, ok := registrar.handler.(func(
		*iwfpb.GetAttributesQueryRequest,
	) (*iwfpb.GetAttributesQueryResponse, error))
	require.True(t, ok)
	response, err := handler(&iwfpb.GetAttributesQueryRequest{AllKeys: true})
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b"}, []string{
		response.GetAttributes()[0].GetKey(),
		response.GetAttributes()[1].GetKey(),
	})
	response, err = handler(nil)
	require.ErrorContains(t, err, "requires a request")
	require.Nil(t, response)
}
