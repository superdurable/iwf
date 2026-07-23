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

func TestStateRecovery(t *testing.T) {
	wfId := "TestStateRecovery" + strconv.Itoa(int(time.Now().Unix()))
	runId, err := client.StartWorkflow(context.Background(), executeApiFailRecoveryWorkflow{}, wfId, 10, nil, nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, runId)

	var output string
	err = client.GetSimpleWorkflowResult(context.Background(), wfId, "", &output)
	assert.Nil(t, err)
	assert.Equal(t, "this is workflow state 2", output)
}
