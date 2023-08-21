# Monitors
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

# How To

## Make a better view of workflow in Temporal UI
To make it easier with iWF, customize the columns by clicking the button from sidebar, to replace WorkflowType with IwfWorkflowType

![Screenshot 2023-05-31 at 3 43 20 PM](https://github.com/indeedeng/iwf/assets/4523955/541a6b93-626f-4bfa-9e0c-924cf8cc7ec9)

## Troubleshoot & Debugging

### Use QueryHandlers: GetDataObjects , GetSearchAttributes,  GetCurrentTimerInfos etc


* GetDataObjects return all the current data attributes of a workflow.  (TODO, rename GetDataObjects to GetDataAttributes )

* GetSearchAttributes return all the current search attributes.

* GetCurrentTimerInfos returns all the current timers.

* PrepareRPCQueryType return all the workflow states' status -- what commands are they waiting for, and what have been completed

### Read Workflow History
Workflow can be used to understand how does a workflow reach to the current status.

**These are ONLY events that you need to understand. You can ignore the others as implementation details of iWF service**

* WorkflowExecutionStarted

It contains the input of the workflow

* ActivityTaskScheduled

It contains the input of State API (waitUntil or execute, depends on the activity type)

* ActivityTaskCompleted

It contains the output of State API (waitUntil or execute, depends on the activity type)

* WorkflowSignaled

If the signal name starts with __IwfSystem_ .. then it's a system signal
They are for internal implementation of iWF service. e.g. SkipTimerChannel is for skipping a timer.
Others are user signals Sent by cient.signalWorkflow API, and received by signalChannelCommand)

* WorkflowExecutionCompleted/Failed

It contains the output of the workflow

