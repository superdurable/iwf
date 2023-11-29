users can customize the WorkflowState

#### WorkflowState WaitUntil/Execute API timeout and retry policy

By default, the API timeout is 30s with infinite backoff retry. 
Users can customize the API timeout and retry policy:

- InitialIntervalSeconds: 1
- MaxInternalSeconds:100
- MaximumAttempts: 0
- MaximumAttemptsDurationSeconds: 0
- BackoffCoefficient: 2

Where zero means infinite attempts.

Both MaximumAttempts and MaximumAttemptsDurationSeconds are used for controlling the maximum attempts for the retry
policy.
MaximumAttempts is directly by number of attempts, where MaximumAttemptsDurationSeconds is by the total time duration of
all attempts including retries. It will be capped to the minimum if both are provided.

#### State API failure handling/recovery

By default, the workflow execution will fail when State APIs max out the retry attempts. In some cases that
workflow want to handle the errors differently.

##### Execute API
For Execute API, you can set `PROCEED_TO_CONFIGURED_STATE` as failure policy, with a `ProceededState` configured.
The proceeded state will take the same input from the original failed state.

When the current state fails, instead of failing the workflow, it will go to the ProceededState for recovery.

Then in the recovery(proceeded) state, it's up to state to continue to run the workflow, or fail the workflow. 

Thius failure policy are especially helpful for recovery logic. 

For example, a `DebitState` is making three API calls for the debit operation but failed at the 3rd one. You want to undo the first two. In that case, you can set a `UndoDebitState` as the recovery state for the DebitState. When DebitState fails, instead of failing workflow, it will proceed to `UndoDebitState` to let you undo the first two operations. 


Example in Java SDK:
```java
public class DebitState extends WorkflowState {
    @Override
    public WorkflowStateOptions getStateOptions() {
        return new WorkflowStateOptionsExtension()
                .setProceedOnExecuteFailure(UndoDebitState.class)
                // make sure the retry duration is less than the workflow timeout so that recovery state has a chance to run
                .executeApiRetryPolicy(...); 
    }
   
    @Override
    public StateDecision execute(...){ 
       // make three API calls for a debit operation
    }
}
```

In Golang SDK:
```golang
type debitState struct{
    iwf.WorkflowStateDefaultsNoWaitUntil
}

func (b debitState) GetStateOptions() *iwf.StateOptions {
	return &iwf.StateOptions{
           // make sure the retry duration is less than the workflow timeout so that recovery state has a chance to run
	options.
           ExecuteApiRetryPolicy: &iwfidl.RetryPolicy{...} 
           ExecuteApiFailureProceedState: undoDebitState{},
        }
}


func (b debitState) Execute(ctx iwf.WorkflowContext, input iwf.Object, commandResults iwf.CommandResults, persistence iwf.Persistence, communication iwf.Communication) (*iwf.StateDecision, error) {
       // make three API calls for a debit operation
}

```

In Python SDK:
```python
class DebitState(WorkflowState[None]):
    def execute(
        self,
        ctx: WorkflowContext,
        input: T,
        command_results: CommandResults,
        persistence: Persistence,
        communication: Communication,
    ) -> StateDecision:
        # make three API calls for a debit operation
        return StateDecision...

    def get_state_options(self) -> WorkflowStateOptions:
        return WorkflowStateOptions(
            execute_api_retry_policy=RetryPolicy(...), //make sure recoveryState has enough time to run
            execute_failure_handling_state=UndoDebitState,
        )
```

##### WaitUntil API

For WaitUntil API, using `PROCEED_ON_API_FAILURE` for `WaitUntilApiFailurePolicy` will let workflow continue to invoke `execute`
API when the API fails with maxing out all the retry attempts.

See example here in [Java](https://github.com/indeedeng/iwf-java-sdk/blob/main/src/test/java/io/iworkflow/integ/basic/ProceedOnStateStartFailWorkflowState1.java#L45) and [Golang](https://github.com/indeedeng/iwf-golang-sdk/blob/main/integ/proceed_on_state_start_fail_workflow_state1.go#L36).

This is very uncommonly needed than the failure policy of Execute API. Currently not implemented in Python SDK yet.

#### State/RPC API Context
There is a context object when invoking RPC or State APIs. It contains information like workflowId, startTime, etc.

For example, WorkflowState can utilize `attempts` or `firstAttemptTime` from the context to make some advanced logic.
