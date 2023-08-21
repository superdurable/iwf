Some tips for making decisions:

* if there will be only one state needs the value, then it’s easier to use state input
* if it’s more than one, you probably want to use data or search attributes to save the code of passing values around 
* By default , all state api / RPC will load all data attributes unless you set a different persistence loading policy. So Having too many data attributes could be wasteful for communication for default loading, and could exceed the 2MB limit . Otherwise you will have to use PartialLoading policy to selectively load data into states. 
* You can set search attributes in the start workflow api without a starting state(starting state is not required, not even a state ) . However, data attributes are not supported in the start api yet. We may support data attributes as well in the 
* StateExecutionLocal is only for passing data from the waitUntil to execute API, within a stateExecutionLocal. This is just for this limited use case. 


Basically if you can use input and it’s easier, then use input. Or if you can use stateExecutionLocal, then use it without hesitation. Or none of them works well, then use persistence. 