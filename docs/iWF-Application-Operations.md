## State API availability 
* Meaning: The State API is failing. The API is implemented by the iWF application. 
* Investigation:
  * Look at logs to see why the State API is failing.
  * If there are no error logs, then check if there is some devops/network issue between iWF server and application 
  * If you can find the workflow in Temporal/Cadence WebUI, look at the pending activity tab to see the error details
  * The failure eventually will be recroded in AcitivytTaskFailure event after max out backoff retry
  * Also use history events to trouble shoot why a workflow has reach to a certain state

## Workflow failure 
* Meaning: There are two cases that workflow could fail:
* Investigation: look at the failed workflows from Temporal UI ( search for ExecutionStatus="Failed" in the UI ):
  * User WorkflowState decides to fail in their code. It’s user behavior. User workflow max out backoff retry in State APIs
  * See the above State API availability alert runbook
  * Mitigation: reset the failed workflows if the bug got fixed and you want to mitigate. See more in HowTo section.

## Workflow timeout
* Meaning: workflow exceed the expected timeout period. The timeout is set when starting the workflow. (Note that users can set 0 for infinite timeout)
* Investigation: look at the failed workflows from Temporal Web UI ( search for ExecutionStatus="TimedOut" in the UI
  * To see the timeout value when starting the workflow, you have to look at the raw JSON history, in workflowExecutionTimeout field of the first event(WorkflowExecutionStarted) 
  * [This can be improved by this feature](https://github.com/temporalio/ui/issues/1429)
* Mitigation: reset the failed workflows if the bug got fixed and you want to mitigate. See more in HowTo section.