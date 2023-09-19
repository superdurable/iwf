Similar to https://github.com/indeedeng/iwf/wiki/How-to-wait-for-a-workflow-to-complete 

External application and make a blocking API call to wait for a state execution to complete.

To wait for a workflow state execution to complete, the stateExecutionId must be provided as `waitForCompletionStateExecutionIds` in the startWorkflowOptions.

Then use `client.waitForStateExecution(stateExecutionId)` API to wait for the completion. 

Note that this feature cannot be used when the same workflowId reused to start more than one execution. See https://github.com/indeedeng/iwf/issues/349 for future redesign. 