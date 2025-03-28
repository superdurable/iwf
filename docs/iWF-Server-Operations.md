# How to Deploy 

## Option 1: Build & Run

* Run `make bins` to build the binary `iwf-server`
* Make sure you have registered the system search attributes required by iWF server:
    * Keyword: IwfWorkflowType
    * Int: IwfGlobalWorkflowVersion
    * Keyword: IwfExecutingStateIds
    * See [Contribution](./CONTRIBUTING.md) for more detailed commands.
    * For Cadence without advancedVisibility enabled,
      set [executingStateIdMode](https://github.com/indeedeng/iwf/blob/main/config/development_cadence.yaml#L9)
      to DISABLED
* Then run  `./iwf-server start` to run the service . This defaults to serve workflows APIs with Temporal interpreter
  implementation. It requires to have local Temporal setup. See Run with local Temporal.
* Alternatively, run `./iwf-server --config config/development_cadence.yaml start` to run with local Cadence.

## Option 2: Using docker image
You can use docker image to deploy in K8s cluster

You can provide a volume override for this [config](https://github.com/indeedeng/iwf/blob/main/config/config_template.yaml) using the path: `/iwf/config/config_template.yaml` so that you can connect to any Cadence/Temporal cluster.

# Configuration
All the server configuration is defined [here](https://github.com/indeedeng/iwf/blob/main/config/config.go).

# Suggested Monitors and Runbook

## Scale up horizontally
Most of the time, you can just simply add instances when you see cpu/memory got too hot.

For some really, really large use cases, you will need to adjust the number of partitions for Temporal task queue (by default it’s 4). This is needed for ~>400 tasks per second (400 is a rough number, depending on your instance perf)

Reference to temporal [task queue partitions](https://docs.temporal.io/task-queue) and [task queue scalability](https://community.temporal.io/t/scaling-task-queues-correlation-between-numhistoryshards-taskqueuepartitions-number-of-matching-service-hosts/6909/2).

## Service API availability
* Meaning: The iWF API is directly calling Temporal /Cadence service, there are no other dependencies
* Investigation: When something is wrong(low availability or high latency), check the Temporal/Cadence Healthcheck dashboard, and client API to confirm that.
  * If the error/latency is from Temporal API, then contact the Temporal team. Otherwise, it probably is our infrastructure issues.
  * Scalability(need more instance/CPU/Memory)

## Service API Latency
This is similar to Service API availability above 

## Interpreter Worker availability 
* Meaning: iWF Cadence/Temporal worker should be always polling tasks even if there is no task to execute. This means workers are not polling tasks from the Cadence/Temporal service.
* Investigation: Check why the workers are not polling:
  * Worker is crashing(cannot start, bad code deployment or config change)
  * Worker is having trouble connecting to Cadence/Temporal(e.g. Cadence/Temporal healthcheck, infra, network
  * Cadence/Temporal API is failing 
  * Check logs 
  * Check Cadence/Temporal RPC API dashboard

## Workflow task execution failure
* Meaning: This should be very rare. When happens, it means something goes wrong in either interpreter workflow code, or SDK
* Investigation: 
  * Get the workflowId from logs
  * Go to Cadence/Temporal WebUI to see why the workflow task is failing 
  * Check if the workflow is trying to return too big payload to history
  * See [size limit in Search attributes](https://docs.temporal.io/visibility#search-attribute)) and [other limits ](https://docs.temporal.io/kb/temporal-platform-limits-sheet)(2MB limit on activity/memo etc).  
  * If so, ask the application to fix their code and do not return such big payload
  * Check if the workflow history is too long that causes workflow task failing because replay takes too much time
  * Note that this shouldn't happen because iWF does "auto-continueAsNew" to workflow to keep history short. Application could override this config in start workflow API, or [update workflow config API](https://github.com/indeedeng/iwf-idl/blob/77e3713a46e8875707030da729dad4fbceb097f1/iwf.yaml#L242)
  * Also if there are too fast signals/RPC which could be a problem that iWF service doesn't have a chance to continueAsNew
  * This is an potential improvement in [iWF](https://github.com/indeedeng/iwf/issues/236) and [Temporal](https://github.com/temporalio/temporal/issues/4137)
  * For this case, we have to ask application to stop sending too fast signal. Alternatively, implement the rejection in iWF service like the Github issue described
  * Check if recent iWF-service deployment could cause bugs
  * Check if recently SDK upgrade which includes bugs 
  * Check if it's non-deterministic error (NDE)
  * If so, then download the history, and use local IDE to replay to debug why the workflow tasks fail. 
* Mitigation:
  * If it takes too much time to investigate and we need to mitigate, then find the workflows(using Kibana log search) and reset workflows  to see if the failing workflow tasks can succeed. 
  * Alternatively, terminate the problematic workflows so that it won't cause issues for other normal workflows

## Workflow Task scheduled to start latency
* Meaning: This means the workflow task got scheduled, but take a while for the worker to process. The workers are supposed to process tasks immediately (e.g. within a second). 
* Investigation:
  * Check logs if there is some workflow tasks failing causing a lot of retry of workflow tasks. 
  * Find out the workflowId from logs, then go to WebUI to investigate. If workflows are failing, then see above instructions to fix the workflow task failure first
  * Check if there is a spike of the traffic recently and overload the worker capacity causing iwf-service instances to hot. Check CPU/memory. 
  * Probably we need to scale up our worker – add more CPU/memory/instance to the iWF service
Contact Cadence/Temporal need to scale up their cluster to process more tasks in the taskqueue (they need to add more partition to the task queue)

## Workflow task execution latency
* Meaning: Workflow tasks are supposed to execute within 10s. 
* Investigation:
  * Check if CPU/memory is too hot. If so scale up the worker fleet
  * Check logs to see if we can find the workflow tasks that have high latency
  * Check if there is high latency in the Cadence/Temporal API. Executing workflow tasks may need to pull the history to replay. If workflow history is too large, it could take more time for workflow to get the history from Cadence/Temporal. See above instructions for "workflow task execution failure".

