## TL; DR:

SignalChannel can be replaced by internal channel + RPC. But this replacement come with a slight overhead of perf(latency) + cost (two Temporal Cloud actions vs one).

As recommendation, if signal is enough I think we should just use signal. but some complex features may need internal channels. E.g. must use an RPC to return some results, or need an RPC call can send multiple internal channel messages atomically which signal cannot achieve

## Details
Signal is created first, it’s mapped directly to Cadence/Temporal’s signal feature. Because of that, you can use Cadence/Temporal WebUI to send signal directly. Because of that, a signal is mapped directly to a Cadence/Temporal history event of “WorkflowExecutionSignaled” which is very nice to read.

There are two reasons that we introduced InternalChannel later:
* Signal is like sending message to MessageQueue without response. Sometimes users do want to get some results as part of the write operation. So RPC is created. In order to let RPC to wake up some WorkflowStates running in background, we need some message channel.
* The WorkflowStates executions(especially when executing in parallel) may need some synchronization. E.g. 1->2,3,4->5 where 2,3,4 need to be completed together before 5 is executed. This is an internal communication as well. so we call it “internal channel”.

## See more
* https://github.com/indeedeng/iwf/wiki/RPC#signal-channel-vs-rpc
* https://github.com/indeedeng/iwf/wiki/WorkflowState#internalchannel-async-message-queue