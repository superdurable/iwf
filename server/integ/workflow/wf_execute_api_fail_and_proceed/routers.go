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

package wf_execute_api_fail_and_proceed

import (
	"github.com/superdurable/iwf/integ/helpers"
	"github.com/superdurable/iwf/service"
	"log"
	"net/http"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/integ/workflow/common"
)

/**
 * This test workflow has one state, using REST controller to implement the workflow directly.
 *
 * State1:
 *		- WaitUntil method is skipped
 *      - Execute method will intentionally fail
 * StateRecover:
 *		- Execute method will gracefully complete workflow
 */
const (
	WorkflowType      = "wf_execute_api_fail_and_proceed"
	State1            = "S1"
	StateRecover      = "Recover"
	InputData         = "test-data"
	InputDataEncoding = "test-encoding"
)

type handler struct {
	invokeHistory sync.Map
}

func NewHandler() common.WorkflowHandler {
	return &handler{
		invokeHistory: sync.Map{},
	}
}

func (h *handler) ApiV1WorkflowStateStart(c *gin.Context, t *testing.T) {
	helpers.FailTestWithErrorMessage("should not get here", t)
}

func (h *handler) ApiV1WorkflowStateDecide(c *gin.Context, t *testing.T) {
	var req iwfidl.WorkflowStateDecideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("received state decide request, ", req)

	if req.GetWorkflowType() == WorkflowType {
		if value, ok := h.invokeHistory.Load(req.GetWorkflowStateId() + "_decide"); ok {
			h.invokeHistory.Store(req.GetWorkflowStateId()+"_decide", value.(int64)+1)
		} else {
			h.invokeHistory.Store(req.GetWorkflowStateId()+"_decide", int64(1))
		}
	}
	input := req.StateInput
	if req.WorkflowStateId == State1 {
		if input.GetData() == InputData && input.GetEncoding() == InputDataEncoding {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "test-error"})
		} else {
			helpers.FailTestWithErrorMessage("input is not correct: "+input.GetData()+", "+input.GetEncoding(), t)
		}

		return
	}
	if req.WorkflowStateId == StateRecover {
		if input.GetData() == InputData && input.GetEncoding() == InputDataEncoding {
			// Move to completion
			c.JSON(http.StatusOK, iwfidl.WorkflowStateDecideResponse{
				StateDecision: &iwfidl.StateDecision{
					NextStates: []iwfidl.StateMovement{
						{
							StateId: service.GracefulCompletingWorkflowStateId,
						},
					},
				},
			})
		} else {
			helpers.FailTestWithErrorMessage("input is not correct: "+input.GetData()+", "+input.GetEncoding(), t)
		}
		return
	}

	helpers.FailTestWithErrorMessage("should not get here", t)
}

func (h *handler) GetTestResult() (map[string]int64, map[string]interface{}) {
	invokeHistory := make(map[string]int64)
	h.invokeHistory.Range(func(key, value interface{}) bool {
		invokeHistory[key.(string)] = value.(int64)
		return true
	})
	return invokeHistory, nil
}
