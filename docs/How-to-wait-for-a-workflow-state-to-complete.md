Similar to https://github.com/indeedeng/iwf/wiki/How-to-wait-for-a-workflow-to-complete 

External application and make a blocking API call to wait for a state execution to complete.

To wait for a workflow state execution to complete, the stateExecutionId must be provided as `waitForCompletionStateExecutionIds` in the startWorkflowOptions.

Then use `client.waitForStateExecution(stateExecutionId)` API to wait for the completion. 

See this [example](https://github.com/indeedeng/iwf-java-sdk/blob/main/src/test/java/io/iworkflow/integ/TimerTest.java#L33) in Java integ test.

Note1: the state that is waited for, must be registered to the workflowOptions on startWorkflow API. This is a limitation until https://github.com/indeedeng/iwf/issues/349

Note2: that currently, this feature cannot be used with reusing workflowId for more than one executions. See https://github.com/indeedeng/iwf/issues/404 for improvement which require a new field in Temporal API: https://github.com/temporalio/temporal/issues/6348