<!--markdownlint-disable-->


# CRDT

CRDT stands for Conflict free Replicated Data Type and they enable multiple replicas of data to be updated independently and concurrently without the need for complex synchronization protccols


CRDTs can be classified into two main categories based on how they handle concurrent updates and ensure eventual consistency


1) Operation Based CRDT

In operation based CRDTs, each update to the data structure is represented as an operation. Operations are commutative and idempotent, meaning they can be applied in any order and multiple times without chaning the result


2) State based CRDT.

In state based CRDT, each replica maintains its state independently and periodically exchanges its state with other replicas to ensure eventual convergence. CRDTs of this kind use a merge function to combine states of different replicat in a deterministic manner
