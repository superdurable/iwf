## Client

### StartWorkflow 
Starts a workflow 
* Parameters 
  * workflowDefinition
  * workflowID
  * timeoutSeconds 
  * workflowInput
  * workflowOptions 
* Known error handling
  * WorkflowAlreadyStartedError -- when there is a workflowID conflict and cannot start based on "IdReusePolicy"


### WaitForWorkflowCompletion
Make a long poll API to wait for workflow completion 
parameters
workflowID
valueType
error handling
UncompletedError
NotExistsError
LongPollTimeout
### SignalWorkflow
send a message to the signalChannel of a workflow 
parameters
workflowDefinition 
workflowID
channelName
value
error handling
No running workflow 

### ResetWorkflow
Reset a workflow execution to the previous point of the history 
Parameters 
workflowID
reseTypeAndOptions
error handling
Not found

### StopWorkflow
Stop a workflow execution 
Parameters 
workflowID
StopWorkflowOptions
error handling
No running workflow 

### GetDataAttributes or GetSearchAttributes
Read the data or search attributes of a workflow execution
 parameters
workflowDefinition 
workflowID
keys 
error handling
Not found

### SetDataAttributes or SetSearchAttributes
Write the data or search attributes of a workflow execution
 parameters
workflowDefinition 
workflowID
key and value pairs of attributes
error handling
No running workflow 

### SearchWorkflow
Find the workflow execution given a search attribute query
 parameters
query
pageSize
paginationToken 

### InvokeRPC
invoke an RPC of a workflow execution 
Parameters 
workflowID
rpc method (by name of stub)
error handling
No applicable workflow 

### DescribeWorkflow
Get basic info of a workflow execution, including running status 
Parameters 
workflowID
error handling
No applicable workflow 

### SkipTimer
Skip the timer, to let it fire immediately 
Parameters 
workflowID
stateId
stateExecutionNumber : optional
timerCommandID: optional
error handling
No applicable workflow 

### WaitForStateExecutionCompletion
Make long poll API call to wait for a state execution to complete
Parameters 
workflowID
stateId
stateExecutionNumber : optional
waitForKey: optional
Error handling:
LongPollTimeout
Not applicable workflow 

## UnregisteredClient 
UnregisteredClient is the raw client without workflow registry. 

It's useful for calling Client APIs in a different repo that doesn’t host the workflows (which may require more dependencies). Underlying, Client is built on top of UnregisteredClient. 

## More APIs from server

* Trigger continue as new 
Update wf config 

TODO: https://github.com/indeedeng/iwf-java-sdk/issues/285 add them to Client/UnregisteredClient

## ClientOptions 
ClientOptions allows customize a Client for communicating with server, including
* ServerURL
* WorkerURL
* ObjectEncoder
* LongPollTimeout
* RequestHeaders
* RetryConfig -- for local retry on calling server API


