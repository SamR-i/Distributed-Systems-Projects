package chandy_lamport

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

// ​ Max random delay added to packet delivery​​​
const maxDelay = 5

// Simulator is the entry point to the distributed snapshot application.​​
//
// It is a discrete time simulator, i.e. events that happen​ at time t + 1 come
// strictly after events that happen at time t. At each time step, the simulator
// examines messages queued up across all the links in the system and decides
// which ones to deliver to the destination.
//
// The simulator is responsible for starting the snapshot process, inducing servers
// to pass tokens to each other, and collecting the snapshot state after the process
// has terminated.
type Simulator struct {
	time           int
	nextSnapshotId int
	servers        map[string]*Server // key = server ID
	logger         *Logger
	// TODO: ADD MORE FIELDS HERE
	finishedServers map[int]map[string]bool
	allFinishedSnap map[int]bool
}

func NewSimulator() *Simulator {
	return &Simulator{
		0,
		0,
		make(map[string]*Server),
		NewLogger(),
		make(map[int]map[string]bool),
		make(map[int]bool),
	}
}

// Return the receive time of a message after adding a random delay.
// Note: since we only deliver one message to a given server at each time step,
// the message may be received *after* the time step returned in this function.
func (sim *Simulator) GetReceiveTime() int {
	return sim.time + 1 + rand.Intn(5)
}

// Add a server to this simulator with the specified number of starting tokens
func (sim *Simulator) AddServer(id string, tokens int) {
	server := NewServer(id, tokens, sim)
	sim.servers[id] = server
}

// Add a unidirectional link between two servers
func (sim *Simulator) AddForwardLink(src string, dest string) {
	server1, ok1 := sim.servers[src]
	server2, ok2 := sim.servers[dest]
	if !ok1 {
		log.Fatalf("Server %v does not exist\n", src)
	}
	if !ok2 {
		log.Fatalf("Server %v does not exist\n", dest)
	}
	server1.AddOutboundLink(server2)
}

// Run an event in the system
func (sim *Simulator) InjectEvent(event interface{}) {
	switch event := event.(type) {
	case PassTokenEvent:
		src := sim.servers[event.src]
		src.SendTokens(event.tokens, event.dest)
	case SnapshotEvent:
		sim.StartSnapshot(event.serverId)
	default:
		log.Fatal("Error unknown event: ", event)
	}
}

// Advance the simulator time forward by one step, handling all send message events
// that expire at the new time step, if any.
func (sim *Simulator) Tick() {
	sim.time++
	sim.logger.NewEpoch()
	// Note: to ensure deterministic ordering of packet delivery across the servers,
	// we must also iterate through the servers and the links in a deterministic way
	for _, serverId := range getSortedKeys(sim.servers) {
		server := sim.servers[serverId]
		for _, dest := range getSortedKeys(server.outboundLinks) {
			link := server.outboundLinks[dest]
			// Deliver at most one packet per server at each time step to
			// establish total ordering of packet delivery to each server
			if !link.events.Empty() {
				e := link.events.Peek().(SendMessageEvent)
				if e.receiveTime <= sim.time {
					link.events.Pop()
					sim.logger.RecordEvent(
						sim.servers[e.dest],
						ReceivedMessageEvent{e.src, e.dest, e.message})

					// added code for us
					logEventObj := LogEvent{
						serverId:     server.Id,
						serverTokens: server.Tokens,
						event:        ReceivedMessageEvent{e.src, e.dest, e.message},
					}

					fmt.Println(logEventObj.String())
					// added code ends

					sim.servers[e.dest].HandlePacket(e.src, e.message)
					break
				}
			}
		}
	}
}

// Start a new snapshot process at the specified server
func (sim *Simulator) StartSnapshot(serverId string) {
	fmt.Println("Starting simulator snapshost")
	snapshotId := sim.nextSnapshotId
	sim.finishedServers[snapshotId] = make(map[string]bool)
	sim.nextSnapshotId++
	sim.logger.RecordEvent(sim.servers[serverId], StartSnapshot{serverId, snapshotId})
	//TODO
	sim.servers[serverId].StartSnapshot(snapshotId)

}

// Callback for servers to notify the simulator that the snapshot process has
// completed on a particular server
// Correcting the `NotifySnapshotComplete` method in simulator.go
func (sim *Simulator) NotifySnapshotComplete(serverId string, snapshotId int) {
	sim.logger.RecordEvent(sim.servers[serverId], EndSnapshot{serverId, snapshotId})
	sim.finishedServers[snapshotId][serverId] = true
	allFinished := true

	for serverId, _ := range sim.servers {
		allFinished = allFinished && sim.finishedServers[snapshotId][serverId] 
	}
	sim.allFinishedSnap[snapshotId] = allFinished
	
	if allFinished {
		fmt.Println("Simulator snapshot complete")
	}
}

// Collect and merge snapshot state from all the servers.
// This function blocks until the snapshot process has completed on all servers.
func (sim *Simulator) CollectSnapshot(snapshotId int) *SnapshotState {
    result := make(chan *SnapshotState) // Create a channel to send the snapshot

    go func() {
        for !sim.allFinishedSnap[snapshotId] {
            fmt.Println("In loop")
            time.Sleep(100 * time.Millisecond)
        }

        fmt.Println("Collecting snapshot")
        snap := &SnapshotState{snapshotId, make(map[string]int), make([]*SnapshotMessage, 0)}

        for serverId, server := range sim.servers {
            fmt.Printf("Processing server %s\n", serverId)
            serverState := server.stateMap[snapshotId]
            recordedMessages := server.recordedMessages[snapshotId]

            // Store the collected server state (e.g., number of tokens).
            snap.tokens[serverId] = serverState

            // Store all recorded messages.
            for mSrc, messages := range recordedMessages {
                for _, msg := range messages {
                    fmt.Printf("Message src %s dest %s: %v\n", mSrc, serverId, msg)
                    snapMsg := &SnapshotMessage{
                        src:     mSrc,
                        dest:    serverId,
                        message: msg,
                    }
                    snap.messages = append(snap.messages, snapMsg)
                }
            }
        }

        for i, count := range snap.tokens {
            fmt.Printf("Server %s: %d\n", i, count)
        }

        fmt.Printf("Snapshot %v\n", snap)
        result <- snap // Send the snapshot through the channel
    }()

    return <-result // Return the snapshot received from the channel
}

