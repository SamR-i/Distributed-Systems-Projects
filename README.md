# Distributed Systems Algorithms Implementations

This repository contains implementations of three fundamental algorithms in distributed systems: MapReduce, Chandy-Lamport, and Raft. Each project is designed to showcase a specific aspect of distributed computing, from data processing to state consistency and global snapshot capturing.

## Projects Overview

### 1. MapReduce

The MapReduce implementation is a model for processing large data sets with a distributed algorithm, using a cluster to conduct various tasks in parallel. This project aims to demonstrate efficient data processing over large clusters, emphasizing throughput and fault tolerance.

- **Language:** Go
- **Key Features:** Distributed computation, fault tolerance
- **[More Details](./MapReduce/README.md)**

### 2. Chandy-Lamport Algorithm

The Chandy-Lamport algorithm is used for capturing consistent global states of a distributed system asynchronously. It is particularly useful for debugging and failure recovery in distributed applications.

- **Language:** Go
- **Key Features:** Global state capturing, non-intrusive snapshot algorithm
- **[More Details](./Chandy-Lamport/README.md)**

### 3. Raft Consensus Algorithm

The Raft consensus algorithm ensures that all nodes in a distributed network agree on the sequence of log entries in a reliable and efficient manner. It’s tailored for systems that require data consistency and availability even in the face of failures.

- **Language:** Go
- **Key Features:** Leader election, log replication, safety and reliability
- **[More Details](./Raft/README.md)**

## Getting Started

To run any of these projects, you will need to have Go installed on your machine. Each project has its own set of dependencies and setup instructions, which can be found in the respective project directories.

### Prerequisites

- Install [Go](https://golang.org/dl/) (version 1.15 or later recommended)
- Git for cloning the repository

