# Chandy-Lamport Algorithm Implementation

## Project Overview

This project implements the Chandy-Lamport algorithm for distributed snapshots in Go. It allows for capturing a global state of a distributed system asynchronously without pausing the system's operations. This is useful for debugging, monitoring, and recovery purposes in distributed applications.

## Functionality

The Chandy-Lamport algorithm includes several components:

- **Simulator:** Simulates a distributed system where processes can send and receive messages.
- **Logger:** Captures and logs the state of each process and the messages in transit when a snapshot is initiated.
- **Snapshot Initiator:** Can be triggered by any process to start the snapshot process.
- **Server:** Manages message passing and state requests between nodes in the system.

## How It Works

### Snapshot Initiation
Any node (process) can initiate a snapshot at any time. It begins by recording its own state and sends a marker message to all other processes.

### Marker Sending and Receiving
When a process receives a marker:
- If it is the first marker it has seen, it records its state and sends marker messages to all other processes.
- If it has already recorded its state, it records the state of the incoming channel from which the marker was received.

### State Recording
The state of each process and the state of each channel are recorded. Channels are assumed to be empty if no messages are in transit at the time of the marker's arrival.

### Completion
Once all processes have recorded their states and all channel states are recorded, the snapshot is considered complete.

## Running the Project

To run the Chandy-Lamport snapshot algorithm simulation:

1. Start the server to manage nodes and message passing.
2. Initialize the nodes and begin normal operations.
3. Trigger a snapshot from any node to start the state recording process.

## Conclusion

This implementation of the Chandy-Lamport algorithm demonstrates the capability to perform global snapshots in a non-blocking manner. It leverages Go’s concurrency model and networking capabilities to efficiently handle distributed state recording.
