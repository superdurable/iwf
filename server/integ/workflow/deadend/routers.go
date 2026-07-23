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

package deadend

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/integ/helpers"
	"github.com/superdurable/iwf/integ/workflow/common"
	"github.com/superdurable/iwf/service"
	"log"
	"net/http"
	"sync"
	"testing"
)

/**
 * This test workflow has 3 states, using REST controller to implement the workflow directly.
 *
 * RPCWriteData:
 *		- WaitUntil will upsert data attributes
 * RPCTriggerState:
 *		- WaitUntil will move to State1
 * State1:
 *		- WaitUntil is skipped
 *      - Execute method will put the state into a dead-end.
 */
const (
	WorkflowType    = "deadend"
	RPCTriggerState = "test-RPCTriggerState"
	RPCWriteData    = "RPCWriteData"

	State1 = "S1"
)

type handler struct {
	invokeHistory sync.Map
}

func NewHandler() common.WorkflowHandlerWithRpc {
	return &handler{
		invokeHistory: sync.Map{},
	}
}

func (h *handler) ApiV1WorkflowWorkerRpc(c *gin.Context, t *testing.T) {
	var req iwfidl.WorkflowWorkerRpcRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("received workflow worker rpc request, ", req)

	wfCtx := req.Context
	if wfCtx.WorkflowId == "" || wfCtx.WorkflowRunId == "" {
		helpers.FailTestWithErrorMessage("invalid context in the request", t)
	}
	if req.WorkflowType != WorkflowType {
		helpers.FailTestWithErrorMessage("invalid workflow type", t)
	}

	if req.RpcName == RPCTriggerState {
		// Move to State 1
		c.JSON(http.StatusOK, iwfidl.WorkflowWorkerRpcResponse{
			StateDecision: &iwfidl.StateDecision{NextStates: []iwfidl.StateMovement{
				{
					StateId: State1,
					StateOptions: &iwfidl.WorkflowStateOptions{
						SkipStartApi: iwfidl.PtrBool(true),
					},
				},
			}},
		})
	} else if req.RpcName == RPCWriteData {
		// Upsert data attributes
		c.JSON(http.StatusOK, iwfidl.WorkflowWorkerRpcResponse{
			UpsertDataAttributes: []iwfidl.KeyValue{
				{
					Key: iwfidl.PtrString("any key"),
					Value: &iwfidl.EncodedObject{
						Encoding: iwfidl.PtrString("encoding"),
						Data:     iwfidl.PtrString("data"),
					},
				},
			},
		})
	} else {
		helpers.FailTestWithErrorMessage(fmt.Sprintf("invalid rpc name: %s", req.RpcName), t)
	}
}

// ApiV1WorkflowStateStart - for a workflow
func (h *handler) ApiV1WorkflowStateStart(c *gin.Context, t *testing.T) {
	helpers.FailTestWithErrorMessage("should not be called", t)
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

		// Move to the dead-end state
		if req.GetWorkflowStateId() == State1 {

			c.JSON(http.StatusOK, iwfidl.WorkflowStateDecideResponse{
				StateDecision: &iwfidl.StateDecision{
					NextStates: []iwfidl.StateMovement{
						{
							StateId: service.DeadEndWorkflowStateId,
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
