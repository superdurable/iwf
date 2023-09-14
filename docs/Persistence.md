
As writing code with programming model, you must have to deal with _data_ everywhere. 
iWF provides a Key-Value storage out of the box. This eliminates the need to depend on a database to implement your workflow.

Your data are stored as Data Attributes and Search Attributes. Together both define the "persistence schema".
The persistence schema is defined and maintained in the code along with other business logic.

Search Attributes work like infinite indexes in a traditional database. You
only need to specify which attributes should be indexed, without worrying about complications you might be used to in
a traditional database like the number of shards, and the order of the fields in an index.

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

