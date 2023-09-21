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
workflow want to ignore the errors.

For WaitUntil API, using `PROCEED_ON_API_FAILURE` for `WaitUntilApiFailurePolicy` will let workflow continue to invoke `execute`
API when the API fails with maxing out all the retry attempts.

For Execute API, you can use `PROCEED_TO_CONFIGURED_STATE` similarly, but it's required to set the `ExecuteApiFailureProceedStateId` to use with it.
Note that the proceeded state will take the same input from the original failed state.

The failure policies are especially helpful for recovery logic. For example, a workflow state may have errors that you want to eventually do a cleanup/recovery to handle.

#### State/RPC API Context
There is a context object when invoking RPC or State APIs. It contains information like workflowId, startTime, etc.

For example, WorkflowState can utilize `attempts` or `firstAttemptTime` from the context to make some advanced logic.
