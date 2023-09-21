WorkflowState is how you implement your asynchronous process as a "workflow".  
It will run in the background, with infinite backoff retry by default. 
 
A WorkflowState is itself like “a small workflow” of 1 or 2 steps:

**[ `waitUntil` ] → `execute`**

**The `waitUntil` API** returns "[commands](#commands-from-waituntil)" to wait for. When the commands are completed, the `execute` API will be invoked.


The `waitUntil` API is optional. If not defined, then the `execute` API will be invoked immediately when the Workflow State is started.

The `execute` API returns a StateDecision to decide what is next.

Both `waitUntil` and `execute` are implemented by code and executed in runtime dynamically. They are both hosted as REST API for iWF server to call. 
It's extremely flexible for business -- [any code change deployed will take effect immediately](https://github.com/indeedeng/iwf/wiki/How-to-modify-workflow-code-without-breaking-changes). 

### StateDecision from `execute` 
User workflow implements a **`execute` API** to return a StateDecision for:
* A next state
* Multiple next states running in parallel
* Stop the workflow:
  * Graceful complete -- Stop the thread, and also will stop the workflow when all other threads are stopped
  * Force complete -- Stop the workflow immediately
  * Force fail  -- Stop the workflow immediately with failure
* Dead end -- Just stop the thread
* Atomically go to next state with condition(e.g. channel is not empty)

State Decisions let you orchestrate the WorkflowState as complex as needed for any use case!

![StateDecision examples](https://github.com/indeedeng/iwf-java-samples/assets/4523955/83f127c2-42d1-454a-a688-389e5419f2bd)



### Commands from `waitUntil`

iWF provides three types of commands:


* `TimerCommand` -- Wait for a **durable timer** to fire.
* `InternalChannelCommand` -- Wait for a message from InternalChannel.
* ~~`SignalCommand` -- [Legacy, Use InternalChannelCommand + RPC instead]Wait for a signal to be published to the workflow signal channel. External applications can use
  SignalWorkflow API to signal a workflow~~.

The `waitUntil` API can return multiple commands along with a `CommandWaitingType`:

* `AllCommandCompleted` -- Wait for all commands to be completed.
* `AnyCommandCompleted` -- Wait for any of the commands to be completed.
* `AnyCommandCombinationCompleted` -- Wait for any combination of the commands in a specified list to be completed.

### InternalChannel: async message queue


iWF provides message queue called `InternalChannel`. User can just declare it in the workflow code without any management at all.
A message sent to the InternalChannel is persisted on server side, delivered to any WorkflowState that is waiting for it with `waitUntil`. 

Message can be sent to an InternalChannel by a WorkflowState or RPC.

Note that the scope of an InternalChannel is only within its workflow execution (not shared across workflows).

#### Usage 1: Waiting for external event/request
[RPC](#rpc) provides an API as mechanism to external application to interact with a workflow. Within an RPC, it can send a message to the internalChannel.
This allows workflowState to be waiting for an external event/request before proceeding. E.g., a workflow can wait for an approval before updating the database. 

#### Usage 2: Multi-thread synchronization
When there are multiple threads of workflow states running in parallel, you may want to have them wait on each other to ensure some particular ordering.

For example, in your problem space, WorkflowStates 1,2,3 need to be completed before WorkflowState 4. 

In this case, you need to utilize the "InternalChannel". WorkflowState 4 should be waiting on an "InternalChannel" for 3 messages via the `waitUntil` API. 
WorkflowState 1,2,3 will each publish a message when completing. This ensures propper ordering.  

A full execution flow of a single WorklfowState can look like this:

![Workflow State diagram](https://user-images.githubusercontent.com/4523955/234921554-587d8ad4-84f5-4987-b838-959869293465.png)

### SDKs

To implement a WorkflowState, just implement the:
* [Java interface](https://github.com/indeedeng/iwf-java-sdk/blob/main/src/main/java/io/iworkflow/core/WorkflowState.java)
* [Golang interface](https://github.com/indeedeng/iwf-golang-sdk/blob/main/iwf/workflow_state.go)
* [Python Base Class](https://github.com/indeedeng/iwf-python-sdk/blob/main/iwf/workflow_state.py)

#### Java
For Java, the `waitUntil` has a default implementation so you just not implement it, and SDK will skip it to invoke `execute` directly.


A full Java WorkflowState looks like:
```java
class WaitSignalOrTimerState implements WorkflowState<Void> {

    @Override
    public Class<Void> getInputType() {
        return Void.class;
    }

    @Override
    public CommandRequest waitUntil(final Context context, final Void input, final Persistence persistence, final Communication communication) {
        return CommandRequest.forAnyCommandCompleted(
                TimerCommand.createByDuration(Duration.ofHours(24)),
                SignalCommand.create(READY_SIGNAL)
        );
    }

    @Override
    public StateDecision execute(final Context context, final Void input, final CommandResults commandResults, final Persistence persistence, final Communication communication) {
        if (commandResults.getAllTimerCommandResults().get(0).getTimerStatus() == TimerStatus.FIRED) {
            return StateDecision.singleNextState(State4.class);
        }
        
        String someData = persistence.getDataAttribute(DA_DATA1, String.class);
        System.out.println("call API3 with backoff retry in this method..");
        return StateDecision.gracefulCompleteWorkflow();
    }
}
```
#### Golang

Golang interface doesn't have default implementation. As a result, put `iwf.WorkflowStateDefaultsNoWaitUntil` into the struct to skip `waitUntil`.

```golang
type state1 struct {
	iwf.WorkflowStateDefaultsNoWaitUntil
}
```

But if it needs waitUntil:

```golang
type state1 struct {
	iwf.WorkflowStateDefaults
}
```

For Golang a full state is like:
```golang
type state3 struct {
	iwf.WorkflowStateDefaults
	svc service.MyService
}

func (i state3) WaitUntil(ctx iwf.WorkflowContext, input iwf.Object, persistence iwf.Persistence, communication iwf.Communication) (*iwf.CommandRequest, error) {
	return iwf.AnyCommandCompletedRequest(
		iwf.NewTimerCommand("", time.Now().Add(time.Hour*24)),
		iwf.NewSignalCommand("", SignalChannelReady),
	), nil
}

func (i state3) Execute(ctx iwf.WorkflowContext, input iwf.Object, commandResults iwf.CommandResults, persistence iwf.Persistence, communication iwf.Communication) (*iwf.StateDecision, error) {
	var data string
	persistence.GetDataAttribute(keyData, &data)
	i.svc.CallAPI3(data)

	if commandResults.Timers[0].Status == iwfidl.FIRED {
		return iwf.SingleNextState(state4{}, nil), nil
	}
	return iwf.GracefulCompletingWorkflow, nil
}
```
#### Python
For Python, the `wait_until` has a default implementation so you just not implement it, and SDK will skip it to invoke `execute` directly.

A full state is like:
```python
class TimerOrInternalChannelState(WorkflowState[None]):
    def wait_until(self, ctx: WorkflowContext, input: T, persistence: Persistence, communication: Communication,
                   ) -> CommandRequest:
        return CommandRequest.for_any_command_completed(
            TimerCommand.timer_command_by_duration(
                timedelta(seconds=10)
            ),  # use 10 seconds for demo
            InternalChannelCommand.by_name(verify_channel),
        )

    def execute(self, ctx: WorkflowContext, input: T, command_results: CommandResults, persistence: Persistence,
                communication: Communication,
                ) -> StateDecision:
        form = persistence.get_data_attribute(data_attribute_form)
        if (
                command_results.internal_channel_commands[0].status
                == ChannelRequestStatus.RECEIVED
        ):
            print(f"API to send welcome email to {form.email}")
            return StateDecision.graceful_complete_workflow("done")
        else:
            print(f"API to send the a reminder email to {form.email}")
            return StateDecision.single_next_state(VerifyState)
```