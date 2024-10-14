
As writing code with programming model, you must have to deal with _data_ everywhere. 
iWF provides a Key-Value storage out of the box. This eliminates the need to depend on a database to implement your workflow.

Your data are stored as Data Attributes and Search Attributes. Together both define the "persistence schema".
The persistence schema is defined and maintained in the code along with other business logic. 

Search Attributes work like infinite indexes in a traditional database. You
only need to specify which attributes should be indexed, without worrying about complications you might be used to in
a traditional database like the number of shards, and the order of the fields in an index. Note that Search Attributes must be created accordingly in the Temporal namespace/Cadence domain, and shared across the whole namespace/domain. 

Logically, the workflow definition displayed in the example workflow diagram will have a persistence schema as follows:

| Workflow Execution   | Search Attr A | Search Attr B | Data Attr C | Data Attr D |
|----------------------|---------------|:-------------:|------------:|------------:|
| Workflow Execution 1 | val 1         |     val 2     |       val 3 |       val 4 |
| Workflow Execution 2 | val 5         |     val 6     |       val 7 |       val 8 |
| ...                  | ...           |      ...      |         ... |         ... |

With Search attributes, you can write [customized SQL-like queries to find any workflow execution(s)](https://docs.temporal.io/visibility#search-attribute), just like using a database query.

Note:
* The scope of the data/search attribute are isolated within its own workflow execution
* Lifecycle: after workflows are closed(completed, timeout, terminated, canceled, failed), all the data retained in your persistence schema will be deleted once the configured retention period elapses.

The iWF persistence is mainly for storing the workflow intermediate states/data.
**It is important to not abuse iWF persistence for things like permanent storage, or for tracking/analytics purpose.**

## Persistence loading policy

When a workflowState/RPC API loads DataAttributes/SearchAttributes, by default it will use `LOAD_ALL_WITOUT_LOCKING` to load everything. There are other options:

* For WorkflowState, there is a 2MB limit by default to load data. User can use another loading policy `LOAD_PARTIAL_WITHOUT_LOCKING`
to specify certain DataAttributes/SearchAttributes only to load.

* `LOAD_NONE` will skip the loading to save the data transportation/history cost.

* `WITHOUT_LOCKING` here means if multiple StateExecutions/RPC try to upsert the same DataAttribute/SearchAttribute, they can be
done in parallel without locking.

* If racing conditions could be a problem, using`PARTIAL_WITH_EXCLUSIVE_LOCK` or `LOAD_ALL_WITH_PARTIAL_LOCK` allow specifying some keys to be locked during the execution.

The locking with RPC is only supported by Temporal as backend with enabling synchronous update feature (for self hosted Temporal cluster, enable by `frontend.enableUpdateWorkflowExecution:true` in Dynamic Config)
See the [wiki](https://github.com/indeedeng/iwf/wiki/What-does-the-atomicity-of-RPC-really-mean%3F) for further details.

## SDKs
Defining iWF persistence schema is simply declaring in code the key and value types(if applicable). 
With the type defined for the attribute, the SDK will check the type matching when read/write. (note that the type enforcement is only on the SDK. The server doesn't care about the types for a data/search attribute -- they are just transparent data blobs.

#### Java


An [example](https://github.com/indeedeng/iwf-java-samples/blob/main/src/main/java/io/iworkflow/workflow/signup/UserSignupWorkflow.java) of Java workflow definition with persistence:
```java
public class UserSignupWorkflow implements ObjectWorkflow {

    public static final String DA_FORM = "Form";

    public static final String DA_Status = "Status";
...

    @Override
    public List<PersistenceFieldDef> getPersistenceSchema() {
        return Arrays.asList(
                DataAttributeDef.create(SignupForm.class, DA_FORM),
                DataAttributeDef.create(String.class, DA_Status)
        );
    }

...
}
```
Example of read/write persistence:
```
        String status = persistence.getDataAttribute(DA_Status, String.class);
        persistence.setDataAttribute(DA_Status, "verified");
```

#### Python

[Example](https://github.com/indeedeng/iwf-python-samples/blob/main/signup/signup_workflow.py) in Python with persistence:
```python
class UserSignupWorkflow(ObjectWorkflow):

    def get_persistence_schema(self) -> PersistenceSchema:
        return PersistenceSchema.create(
            PersistenceField.data_attribute_def(data_attribute_form, Form),
            PersistenceField.data_attribute_def(data_attribute_status, str),
            PersistenceField.data_attribute_def(data_attribute_verified_source, str),
        )
```
Example of read/write persistence:
```
        status = persistence.get_data_attribute(data_attribute_status)
        persistence.set_data_attribute(data_attribute_status, "verified")
```

#### Golang
Due to the limitation of Golang, the Golang SDK doesn't let you define "type" of an attribute. So there is no type checking in the SDK.

This is an [example](https://github.com/indeedeng/iwf-golang-samples/blob/main/workflows/microservices/workflow.go) of a Golang workflow definition with persistence:
```golang
type OrchestrationWorkflow struct {
	iwf.WorkflowDefaults
}

func (e OrchestrationWorkflow) GetPersistenceSchema() []iwf.PersistenceFieldDef {
	return []iwf.PersistenceFieldDef{
		iwf.DataAttributeDef(keyData),
	}
}

func (e OrchestrationWorkflow) GetCommunicationSchema() []iwf.CommunicationMethodDef {
	return []iwf.CommunicationMethodDef{
		iwf.SignalChannelDef(SignalChannelReady),

		iwf.RPCMethodDef(e.MyRPC, nil),
	}
}
```
And example to read/write the persistence:
```golang
func (s *MyState)Execute(...){
	var oldData string
	persistence.GetDataAttribute(keyData, &oldData)
	var newData string
	input.Get(&newData)
	persistence.SetDataAttribute(keyData, newData)
        ...
}
```