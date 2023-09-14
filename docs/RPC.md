RPC stands for "Remote Procedure Call". Allows external systems to interact with the workflow execution.

It's invoked by client, executed in workflow worker, and then respond back the results to client. 

RPC can have access to not only persistence read/write API, but also interact with WorkflowStates using InternalChannel, 
or trigger a new WorkflowState execution in a new thread.

### Atomicity of RPC APIs

It's important to note that in addition to read/write persistence fields, a RPC can **trigger new state executions, and publish message to InternalChannel, all atomically.**

Atomically sending internal channel, or triggering state executions is an important pattern to ensure consistency across dependencies for critical business – this 
solves a very common problem in many existing distributed system applications. Because most RPCs (like REST/gRPC/GraphQL) don't provide a way to invoke 
background execution when updating persistence. People sometimes have to use complicated design to acheive this. 

**But in iWF, it's all builtin, and user application just needs a few lines of code!** 

![flow with RPC](https://user-images.githubusercontent.com/4523955/234930263-40b98ca7-4401-44fa-af8a-32d5ae075438.png)

Note that by default, read and write are atomic separately.
To ensure the atomicity of the whole RPC for read+write, you should use `PARTIAL_WITH_EXCLUSIVE_LOCK` persistence loading policy for the RPC options.
The `PARTIAL_WITH_EXCLUSIVE_LOCK` for RPC is only supported by Temporal as backend with enabling synchronous update feature (by `frontend.enableUpdateWorkflowExecution:true` in Dynamic Config).
See the [wiki](https://github.com/indeedeng/iwf/wiki/What-does-the-atomicity-of-RPC-really-mean%3F) for further details.

### Signal Channel vs RPC

There are two major ways for external clients to interact with workflows: Signal and RPC. 

Historically, signal was created first as the only mechanism for external application to interact with workflow. However, it's a "write only"
which is limited. RPC is the new way and much more powerful and flexible. 

Here are some more details:
* Signal is sent to iWF service without waiting for response of the processing
* RPC will wait for worker to process the RPC request synchronously
* Signal will be held in a signal channel until a workflow state consumes it
* RPC will be processed by worker immediately

![signals vs rpc](https://user-images.githubusercontent.com/4523955/234932674-b0d062b2-e5dd-4dbe-93b5-1b9863acc5e0.png)