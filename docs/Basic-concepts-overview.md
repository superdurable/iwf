A user application defines an ObjectWorkflow by implementing the Workflow interface, in one of the supported languages e.g.
[Java](https://github.com/indeedeng/iwf-java-sdk/blob/main/src/main/java/io/iworkflow/core/ObjectWorkflow.java)
, [Golang](https://github.com/indeedeng/iwf-golang-sdk/blob/main/iwf/workflow.go) , [Python](https://github.com/indeedeng/iwf-python-sdk/blob/main/iwf/workflow.py), or [Typescript/JavaScript](https://github.com/indeedeng/iwf-ts-sdk/blob/main/iwf/src/object-workflow.ts).

An implementation of the interface is referred to as a `WorkflowDefinition` and consists of the components shown below:

| Name                                                    | Description                                                                                                                                       | 
|:--------------------------------------------------------|:--------------------------------------------------------------------------------------------------------------------------------------------------| 
| [WorkflowState](https://github.com/indeedeng/iwf/wiki/WorkflowState)                        | A basic asyn/background execution unit as a "workflow". A State consists of one or two steps: *waitUntil* (optional) and *execute* with retry     |
| [RPC](https://github.com/indeedeng/iwf/wiki/RPC)                                             | API for application to interact with the workflow. It can access to persistence, internal channel, and state execution                            |
| [Persistence](https://github.com/indeedeng/iwf/wiki/Persistence)                             | A Kev-Value storage out-of-box to storing data. Can be accessed by RPC/WorkflowState implementation.                                              |
| [DurableTimer](https://github.com/indeedeng/iwf/wiki/WorkflowState#commands-from-waituntil)                | The waitUntil API can return a timer command to wait for certain time as a durable timer -- it is persisted by server and will not be lost.       |
| [InternalChannel](https://github.com/indeedeng/iwf/wiki/WorkflowState#internalchannel-async-message-queue) | The waitUntil API can return some command for "Internal Channel" -- An internal message queue workflow                                            |
| ~~[Signal Channel](https://github.com/indeedeng/iwf/wiki/RPC#signal-channel-vs-rpc)~~            | Legacy concept and deprecated. Use InternalChannel + RPC instead. A message queue for the workflowState to receive messages from external sources |

