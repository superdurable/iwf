The iWF top level concept is `WorkflowDefinition`which consists of the components shown below:

| Name                                                    | Description                                                                                                                                       | 
|:--------------------------------------------------------|:--------------------------------------------------------------------------------------------------------------------------------------------------| 
| [WorkflowState](https://github.com/indeedeng/iwf/wiki/WorkflowState)                        | A basic asyn/background execution unit as a "workflow". A State consists of one or two steps: *waitUntil* (optional) and *execute* with retry     |
| [RPC](https://github.com/indeedeng/iwf/wiki/RPC)                                             | API for application to interact with the workflow. It can access to persistence, internal channel, and state execution                            |
| [Persistence](https://github.com/indeedeng/iwf/wiki/Persistence)                             | A Kev-Value storage out-of-box to storing data. Can be accessed by RPC/WorkflowState implementation.                                              |
| [DurableTimer](https://github.com/indeedeng/iwf/wiki/WorkflowState#commands-from-waituntil)                | The waitUntil API can return a timer command to wait for certain time as a durable timer -- it is persisted by server and will not be lost.       |
| [InternalChannel](https://github.com/indeedeng/iwf/wiki/WorkflowState#internalchannel-async-message-queue) | The waitUntil API can return some command for "Internal Channel" -- An internal message queue workflow                                            |
| ~~[Signal Channel](https://github.com/indeedeng/iwf/wiki/RPC#signal-channel-vs-rpc)~~            | Legacy concept and deprecated. Use InternalChannel + RPC instead. A message queue for the workflowState to receive messages from external sources |

### SDK

A user application defines an ObjectWorkflow by implementing:
* [Java Interface](https://github.com/indeedeng/iwf-java-sdk/blob/main/src/main/java/io/iworkflow/core/ObjectWorkflow.java)
* [Golang Interface](https://github.com/indeedeng/iwf-golang-sdk/blob/main/iwf/workflow.go) 
* [Python Base Class](https://github.com/indeedeng/iwf-*python-sdk/blob/main/iwf/workflow.py)

The Java interface has default implementation of all methods. So you can skip if you don't need any of them.
For example, if a workflow doesn't need persistence, then just skip the persistenceSchema.

An [example](https://github.com/indeedeng/iwf-java-samples/blob/main/src/main/java/io/iworkflow/workflow/signup/UserSignupWorkflow.java) of Java workflow definition:
```java
public class UserSignupWorkflow implements ObjectWorkflow {

    public static final String DA_FORM = "Form";

    public static final String DA_Status = "Status";
    public static final String VERIFY_CHANNEL = "Verify";

    private MyDependencyService myService;

    public UserSignupWorkflow(MyDependencyService myService) {
        this.myService = myService;
    }

    @Override
    public List<StateDef> getWorkflowStates() {
        return Arrays.asList(
                StateDef.startingState(new SubmitState(myService)),
                StateDef.nonStartingState(new VerifyState(myService))
        );
    }

    @Override
    public List<PersistenceFieldDef> getPersistenceSchema() {
        return Arrays.asList(
                DataAttributeDef.create(SignupForm.class, DA_FORM),
                DataAttributeDef.create(String.class, DA_Status)
        );
    }

    @Override
    public List<CommunicationMethodDef> getCommunicationSchema() {
        return Arrays.asList(
                InternalChannelDef.create(Void.class, VERIFY_CHANNEL)
        );
    }

    // Atomically read/write/send message in RPC
    @RPC
    public String verify(Context context, Persistence persistence, Communication communication) {
        String status = persistence.getDataAttribute(DA_Status, String.class);
        if (status == "verified") {
            return "already verified";
        }
        persistence.setDataAttribute(DA_Status, "verified");
        communication.publishInternalChannel(VERIFY_CHANNEL, null);
        return "done";
    }
}
```

The Python base class has default implementation of all methods. So you can skip if you don't need any of them.
For example, if a workflow doesn't need persistence, then just skip the persistenceSchema.

[Example](https://github.com/indeedeng/iwf-python-samples/blob/main/signup/signup_workflow.py) in Python:
```python
class UserSignupWorkflow(ObjectWorkflow):
    def get_workflow_states(self) -> StateSchema:
        return StateSchema.with_starting_state(SubmitState(), VerifyState())

    def get_persistence_schema(self) -> PersistenceSchema:
        return PersistenceSchema.create(
            PersistenceField.data_attribute_def(data_attribute_form, Form),
            PersistenceField.data_attribute_def(data_attribute_status, str),
            PersistenceField.data_attribute_def(data_attribute_verified_source, str),
        )

    def get_communication_schema(self) -> CommunicationSchema:
        return CommunicationSchema.create(
            CommunicationMethod.internal_channel_def(verify_channel, None)
        )

    @rpc()
    def verify(
            self, source: str, persistence: Persistence, communication: Communication
    ) -> str:
        status = persistence.get_data_attribute(data_attribute_status)
        if status == "verified":
            return "already verified"
        persistence.set_data_attribute(data_attribute_status, "verified")
        persistence.set_data_attribute(data_attribute_verified_source, source)
        communication.publish_to_internal_channel(verify_channel)
        return "done"
```

Golang interface doesn't have default method implementation. So to make it "skippable", you just need to add the default implementation `iwf.DefaultWorkflowType` of all:
```golang
type MyWorkflow struct {
	iwf.DefaultWorkflowType
}
```

Also, Golang doesn't have equivalence to Java's annotation or Python's decorator. An RPC must be registered under CommunicationSchema.

This is an [example](https://github.com/indeedeng/iwf-golang-samples/blob/main/workflows/microservices/workflow.go) of a Golang workflow definition:
```golang
type OrchestrationWorkflow struct {
	iwf.DefaultWorkflowType

	svc service.MyService
}

func (e OrchestrationWorkflow) GetWorkflowStates() []iwf.StateDef {
	return []iwf.StateDef{
		iwf.StartingStateDef(NewState1(e.svc)),
		iwf.NonStartingStateDef(NewState2(e.svc)),
		iwf.NonStartingStateDef(NewState3(e.svc)),
		iwf.NonStartingStateDef(NewState4(e.svc)),
	}
}

func (e OrchestrationWorkflow) GetPersistenceSchema() []iwf.PersistenceFieldDef {
	return []iwf.PersistenceFieldDef{
		iwf.DataAttributeDef(keyData),
	}
}

func (e OrchestrationWorkflow) GetCommunicationSchema() []iwf.CommunicationMethodDef {
	return []iwf.CommunicationMethodDef{
		iwf.SignalChannelDef(SignalChannelReady),

		iwf.RPCMethodDef(e.Swap, nil),
	}
}
```