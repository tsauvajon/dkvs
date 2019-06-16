# Distributed Key-Value store

I was reading **Designing Data-Intensive applications** and wanted to build my
own distributed database.
This is a distributed Key-Value store with basic functionality to better
understand how distributed databases work.

A single master will handle writes and replicate them to n slave nodes, that can
handle reads (the master can also handle reads).

It is far from perfect and definitely not intended for production usage.
Interesting evolutions would be towards reducing inconsistencies, probably
following the path of eventual consistency.

## Functionality

Node:
Can be slave or master.

all nodes, slave or master:
- READ a value
- maintain a list of all nodes
- (Out of scope for this repo) maintain a transaction log or something similar

slave:
- check the master node's health
- elect a new master (while agreeing with other nodes)
- be promoted to master
- join a master:
    - get an ID and a list of nodes from the master
    - update asynchronously until fully up to date with master

master
- check health of the slave nodes
- WRITE value (new or existing)
- push a WRITE query to all slaves
- push an update of the list of nodes when a node joins or leaves 

OUT OF SCOPE:
- database client
- proxy/load balancer to redirect queries to the correct nodes (random or
closest node for reads, master for writes) - this could be included in the client
- security
- authentication

### Design choices

Single write master, many read slaves - no need for a conflict resolver, still
increases availability
Automatic failover by electing a new master when the old one stops responding
for 5 sec (low value used for testing). All un-replicated writes will be lost.

Asynchronous replication on new writes - this could cause inconsistencies if
writes do not get handled in the same order in every node: two successive writes
on the same key, played on different orders on two nodes, would result in
different values in these  nodes  

Async replication when a node joins : all the data is copied from the master to
the joining node by dribs and drabs. This could cause inconsistencies if the
data gets updated while replicating  

Abstract the Storage and Transport layers so it can use different implementations

### Dependencies:

- Generate unique IDs: https://github.com/rs/xid

### References:

- How to handle data in applications:  
[Designing data-intensive applications](https://www.goodreads.com/book/show/23463279-designing-data-intensive-applications)
, *Martin Kleppmann*

- Leader election pattern:  
https://docs.microsoft.com/en-us/azure/architecture/patterns/leader-election

- Leader election algorithm:  
https://www.cs.colostate.edu/~cs551/CourseNotes/Synchronization/BullyExample.html

- Creating a Distributed Hash table:  
https://medium.com/techlog/chord-building-a-dht-distributed-hash-table-in-golang-67c3ce17417b

- Distributed Hash table:  
https://github.com/arriqaaq/chord
