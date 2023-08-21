The page is using Java as example, but should be the same idea for other SDKs.

You can call a Client API after starting workflow, to wait for the completion of the workflow -- getSimpleResultsWithWait  will let you wait for the completion.
Some details about it:

* The API will throw an WorkflowUncompletedException if the workflow doesn’t complete eventually. This include failure, timeout, cancel, terminate etc. You could catch it to do other business.
* It will throw [ClientSideException](https://github.com/indeedeng/iwf-java-sdk/blob/285ca39963cee66f044a89eab9de823117666d9d/src/test/java/io/iworkflow/integ/BasicTest.java#L78C52-L78C82) with WORKFLOW_NOT_EXISTS_SUB_STATUS if the workflow Id is not valid.
* Since it's a long poll API and HTTP request has timeout(default to 60s), it will throw [ClientSideException](https://github.com/indeedeng/iwf-java-sdk/blob/aa6b1667d3b2f35d771883830890436d52709bcb/src/test/java/io/iworkflow/integ/WorkflowUncompletedTest.java#L44) with  
 LONG_POLL_TIME_OUT_SUB_STATUS if exceeds the wait time. You should just retry on that if needs continue waiting. 
* You may just call this API in a State API and let State API keep on retrying for you automatically. 
* You can pass Void.class as result class, If you just need to wait for the result, without any real results to return.


In the future, we will support https://github.com/indeedeng/iwf/issues/268 which will be more flexible for waiting