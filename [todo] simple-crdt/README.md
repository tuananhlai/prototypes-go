## What is a CRDT?

A **CRDT** is a **Conflict-free Replicated Data Type**.

It’s a data structure designed so that multiple copies of the same data can be updated independently on different machines, and later merged **without conflicts**.

The key idea is:

* Each replica can accept writes locally
* Replicas can exchange states later
* Merging is guaranteed to converge to the same result, no matter the order of updates

## Why use it?

CRDTs are useful when you want:

* **Offline-first behavior**
    * Users can keep working without a network connection
* **Low-latency writes**
    * Updates happen locally instead of waiting for a central server
* **Distributed systems**
    * Data is spread across multiple nodes or devices
* **Automatic conflict resolution**
    * You avoid manual conflict handling for every write
* **Eventual consistency**
    * All replicas eventually reach the same state

## Simple intuition

Imagine two phones editing the same counter while disconnected:

* Phone A adds 2
* Phone B adds 1

When they reconnect, instead of arguing about which update “wins,” they merge in a way that preserves both changes.

That’s what CRDTs are for: **safe merging of concurrent updates**.

## When to use them

CRDTs are a good fit for:

* collaborative editing
* chat apps
* counters and likes
* shared todo lists
* distributed caches or sync systems

## Tradeoffs

CRDTs are not always the best choice:

* They can use more memory than simpler approaches
* Some CRDTs are more complex to implement
* Not every data model is easy to express as a CRDT
* They usually give **eventual consistency**, not immediate strong consistency