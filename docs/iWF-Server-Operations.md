Dashboards 
[Server operation Dashboard](TODO)

API Service
Network 
Devops (Marvin)
Scalability(need more instance/CPU/Memory)
The iWF API is directly calling Temporal cloud service, there are no other dependencies. When something is wrong(low availability or high latency), check the Temporal RPC API to confirm that. If the error/latency is from Temporal API, then contact the Temporal team. Otherwise, it probably is our infrastructure issues:
Interpreter Workflow
This section contains detailed info about the status of the interpreter workflow. The metrics are documented here https://docs.temporal.io/references/sdk-metrics/ 
Interpreter Activity
The section contains detailed metrics about the Temporal workflow activity calling iWF workflow worker APIs(WorkflowState start/decide)
Monitor Runbooks

iWF Service Monitor
Service API availability
Meaning: The iWF API is directly calling Temporal cloud service, there are no other dependencies
Investigation: When something is wrong(low availability or high latency), check the Temporal RPC API to confirm that. If the error/latency is from Temporal API, then contact the Temporal team. Otherwise, it probably is our infrastructure issues:
Network
Devops (Marvin)
Scalability(need more instance/CPU/Memory)
Service API Latency
Similar to Service API availability

Worker availability
Meaning: iWF Temporal worker should be always polling tasks even if there is no task to execute. This means workers are not polling tasks from the Temporal service.
NOTE that this monitor is unstable ATM. Still working with Datadog to see why NoData is not reliable to use.
Investigation: Check why the workers are not polling:
Worker is crashing(cannot start, bad code or config change)
Worker is having trouble connecting to Temporal
Temporal API is failing (contact Temporal)

Workflow task execution failure
Meaning: This should be very rare. When happens, it means something goes wrong in either interpreter workflow code, or SDK
Investigation: 
Check if there is new code change recently cause non-deterministic error 
Check if recently SDK upgrade which includes bugs 
Get the workflowId from logs, go to Temporal WebUI to download the history, and use local IDE to replay to debug why the workflow tasks fail. 
Mitigation: If it takes too much time to investigate and we need to mitigate, then find the workflows(using Kibana log search) and reset workflows  to see if the failing workflow tasks can succeed. 

Workflow Task scheduled to start latency
Meaning: This means the workflow task got scheduled, but take a while for the worker to process. The workers are supposed to process tasks immediately (e.g. within a second). 
Investigation:
Check if there is some workflow tasks failing causing a lot of retry of workflow tasks. 
If so, fix the workflow task failure first. 
Check if there is a spike of the traffic recently and overload the worker capacity
Probably we need to scale up our worker – add more CPU/memory/instance to the fleet
If above still doesn’t work, probably Temporal need to scale up their cluster to process more tasks in the taskqueue (they need to add more partition to the task queue)

Workflow task execution latency
Meaning: Workflow tasks are supposed to execute within 10s. Taking too much time means something is off. 
Investigation:
Check if CPU/memory is too hot. If so scale up the worker fleet
Check if there is high latency in the Temporal API. Executing workflow tasks may need to pull the history to replay. If workflow history is too large, it could take more time for workflow to get the history from Temporal. Contact the temporal team to find those workflows. If that’s the case, contact the user to ask them to do continueAsNew in iWF application.