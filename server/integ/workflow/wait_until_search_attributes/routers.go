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

package wait_until_search_attributes

import (
	"github.com/gin-gonic/gin"
	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/integ/workflow/common"
	"github.com/superdurable/iwf/service"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"
)

/**
 * This test workflow has 2 states, using REST controller to implement the workflow directly.
 *
 * State1:
 *		- Waits on nothing. Will execute momentarily
 *      - Execute method will go to State2
 * State2:
 *		- Waits on nothing. Will execute momentarily
 *		- Skips wait until
 *      - 10-second delay is added before executing state
 *      - Execute method will gracefully complete workflow
 */
const (
	WorkflowType = "wait_until_search_attributes"
	State1       = "S1"
	State2       = "S2"

	TestSearchAttributeExecutingStateIdsKey = "IwfExecutingStateIds"
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
	var req iwfidl.WorkflowStateStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("received state start request, ", req)

	if req.GetWorkflowType() == WorkflowType {
		if value, ok := h.invokeHistory.Load(req.GetWorkflowStateId() + "_start"); ok {
			h.invokeHistory.Store(req.GetWorkflowStateId()+"_start", value.(int64)+1)
		} else {
			h.invokeHistory.Store(req.GetWorkflowStateId()+"_start", int64(1))
		}

		if req.GetWorkflowStateId() == State1 {
			// Go straight to the decide methods without any commands
			c.JSON(http.StatusOK, iwfidl.WorkflowStateStartResponse{
				CommandRequest: &iwfidl.CommandRequest{
					DeciderTriggerType: iwfidl.ALL_COMMAND_COMPLETED.Ptr(),
				},
			})
			return
		}
		if req.GetWorkflowStateId() == State2 {
			// Go straight to the decide methods without any commands
			c.JSON(http.StatusOK, iwfidl.WorkflowStateStartResponse{
				CommandRequest: &iwfidl.CommandRequest{
					DeciderTriggerType: iwfidl.ALL_COMMAND_COMPLETED.Ptr(),
				},
			})
			return
		}
	}

	c.JSON(http.StatusBadRequest, struct{}{})
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

		if req.GetWorkflowStateId() == State1 {
			// Move to State 2, skipping wait until
			c.JSON(http.StatusOK, iwfidl.WorkflowStateDecideResponse{
				StateDecision: &iwfidl.StateDecision{
					NextStates: []iwfidl.StateMovement{
						{
							StateId: State2,
							StateOptions: &iwfidl.WorkflowStateOptions{
								SkipWaitUntil: iwfidl.PtrBool(true),
							},
						},
					},
				},
			})
			return
		} else if req.GetWorkflowStateId() == State2 {
			time.Sleep(time.Second * 10)
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
			return
		}

	}

	c.JSON(http.StatusBadRequest, struct{}{})
}

func (h *handler) GetTestResult() (map[string]int64, map[string]interface{}) {
	invokeHistory := make(map[string]int64)
	h.invokeHistory.Range(func(key, value interface{}) bool {
		invokeHistory[key.(string)] = value.(int64)
		return true
	})
	return invokeHistory, nil
}
