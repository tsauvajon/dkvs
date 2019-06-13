Distributed Key-Value store

Node:
Can be slave or master.

all nodes, slave or master:
- GET value
- maintain a list of all nodes
- maintain an ordered list of all transactions
- (Out of scope for this repo) maintain a transaction log or something similar

slave:
- check health of the master node
- elect a new master (while agreeing with other nodes)
- be promoted to master
- join a master:
    - get an ID and a list of nodes from the master
    - update by dribs and drabs until fully up to date with master

master
- check health of the slave nodes
- SET value (new or existing)
- create an ID for the transaction
- push a SET query to all slaves
- push an update to the list of nodes 

TODO:
design a way to proxy the queries:
- should the proxy be embedded in the nodes
- should it be a separate thing
- how to keep an up to date list of nodes


### References:

- How to design a distributed database:  
[Designing data-intensive applications](https://www.goodreads.com/book/show/23463279-designing-data-intensive-applications), *Martin Kleppmann*

- Leader election pattern:  
https://docs.microsoft.com/en-us/azure/architecture/patterns/leader-election

- Leader election algorithm:  
https://www.cs.colostate.edu/~cs551/CourseNotes/Synchronization/BullyExample.html

- Creating a Distributed Hash table:  
https://medium.com/techlog/chord-building-a-dht-distributed-hash-table-in-golang-67c3ce17417b
