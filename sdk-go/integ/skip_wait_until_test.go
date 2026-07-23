// Copyright (c) 2022-2026 Super Durable, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integ

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSkipWaitUntil(t *testing.T) {
	wfId := "TestSkipWaitUntil" + strconv.Itoa(int(time.Now().Unix()))
	runId, err := client.StartWorkflow(context.Background(), &skipWaitUntilWorkflow{}, wfId, 10, 1, nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, runId)

	var output int
	err = client.GetSimpleWorkflowResult(context.Background(), wfId, "", &output)
	assert.Nil(t, err)
	assert.Equal(t, 3, output)
}

func TestSkipWaitUntilWithStruct(t *testing.T) {
	wfId := "TestSkipWaitUntil" + strconv.Itoa(int(time.Now().Unix()))
	runId, err := client.StartWorkflow(context.Background(), skipWaitUntilWorkflow2{}, wfId, 10, 1, nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, runId)

	var output int
	err = client.GetSimpleWorkflowResult(context.Background(), wfId, "", &output)
	assert.Nil(t, err)
	assert.Equal(t, 3, output)
}
