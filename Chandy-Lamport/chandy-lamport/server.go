package chandy_lamport

import (
	"fmt"
	"log"
	"sync"
)

// The main participant of​ the distributed snapshot​ protocol.
// Servers exchange token messages and marker messages among each other.
// Token messages represent the transfer of tokens from one server to another.
// Marker messages represent the progress of the snapshot process. The bulk of
// the distributed protocol is implemented in `HandlePacket` and `StartSnapshot`.
type Server struct {
	Id            string
	Tokens        int
	sim           *Simulator
	outboundLinks map[string]*Link // key = link.dest
	inboundLinks  map[string]*Link // key = link.src
	// TODO: ADD MORE FIELDS HERE

	// These are the extra fields
	receivedMarkers  map[int]map[string]bool          // Markers received from neighbors
	recordedMessages map[int]map[string][]interface{} // Recorded messages from neighbors
	inSnapshot       map[int]bool                     // Flag indicating if the server is in snapshot mode
	snapshotIdMap    map[int]int                      // ID of the current snapshot
	snapshotLock     sync.Mutex               // Lock to synchronize snapshot state access
	stateMap map[int]int   
	blockedInterfaceMap map[int]string 

}

// A unidirectional communication channel between two servers
// Each link contains an event queue (as opposed to a packet queue)
type Link struct {
	src    string
	dest   string
	events *Queue
}

func NewServer(id string, tokens int, sim *Simulator) *Server {
	/*
		return &Server{
			id,
			tokens,
			sim,
			make(map[string]*Link),
			make(map[string]*Link),
		}
	*/
	return &Server{
		Id:               id,
		Tokens:           tokens,
		sim:              sim,
		outboundLinks:    make(map[string]*Link),
		inboundLinks:     make(map[string]*Link),
		receivedMarkers:  make(map[int]map[string]bool),
		recordedMessages: make(map[int]map[string][]interface{}),
		stateMap:  make(map[int]int),
		blockedInterfaceMap: make(map[int]string),
		snapshotIdMap: make(map[int]int),
		inSnapshot: make(map[int]bool),
	}
}

// Add a unidirectional link to the destination server
func (server *Server) AddOutboundLink(dest *Server) {
	if server == dest {
		return
	}
	l := Link{server.Id, dest.Id, NewQueue()}
	server.outboundLinks[dest.Id] = &l
	dest.inboundLinks[server.Id] = &l
}

// Send a message on all of the server's outbound links
func (server *Server) SendToNeighbors(message interface{}) {
	for _, serverId := range getSortedKeys(server.outboundLinks) {
		link := server.outboundLinks[serverId]
		server.sim.logger.RecordEvent(
			server,
			SentMessageEvent{server.Id, link.dest, message})
		link.events.Push(SendMessageEvent{
			server.Id,
			link.dest,
			message,
			server.sim.GetReceiveTime()})
	}
}

// Send a number of tokens to a neighbor attached to this server
func (server *Server) SendTokens(numTokens int, dest string) {
	if server.Tokens < numTokens {
		log.Fatalf("Server %v attempted to send %v tokens when it only has %v\n",
			server.Id, numTokens, server.Tokens)
	}
	message := TokenMessage{numTokens}
	server.sim.logger.RecordEvent(server, SentMessageEvent{server.Id, dest, message})
	// Update local state before sending the tokens
	server.Tokens -= numTokens
	link, ok := server.outboundLinks[dest]
	if !ok {
		log.Fatalf("Unknown dest ID %v from server %v\n", dest, server.Id)
	}
	link.events.Push(SendMessageEvent{
		server.Id,
		dest,
		message,
		server.sim.GetReceiveTime()})
}

// Callback for when a message is received on this server.
// When the snapshot algorithm completes on this server, this function
// should notify the simulator by calling `sim.NotifySnapshotComplete`.
// HandlePacket processes incoming messages, manages token updates, and snapshot operations.
func (server *Server) HandlePacket(src string, message interface{}) {
	server.snapshotLock.Lock()
	defer server.snapshotLock.Unlock()
	fmt.Printf("%v", server.snapshotIdMap)
		switch msg := message.(type) {
		case TokenMessage:
			for snapID, _ := range server.snapshotIdMap{
				if server.inSnapshot[snapID] && src != server.blockedInterfaceMap[snapID]{
					fmt.Println("New message!!!!!")
					server.recordedMessages[snapID][src] = append(server.recordedMessages[snapID][src], message)
					fmt.Printf("Snapshot: %d: %v\n", snapID, server.recordedMessages[snapID])
				}
			}
			// Increase token count upon receiving tokens.
			server.Tokens += msg.numTokens
			fmt.Printf("Server %s has now %d\n", server.Id, server.Tokens)

			server.sim.logger.RecordEvent(server, ReceivedMessageEvent{server.Id, src, msg.numTokens})
			
		case MarkerMessage:
			// If this is the first marker from any source, start snapshot.
			if !server.inSnapshot[msg.snapshotId] {
				server.StartSnapshot(msg.snapshotId)
				server.blockedInterfaceMap[msg.snapshotId] = src
			}

			// Record the receipt of this marker only if not already received.
			if _, exists := server.receivedMarkers[msg.snapshotId][src]; !exists {
				server.receivedMarkers[msg.snapshotId][src] = true
				//server.recordedMessages[src] = []interface{}{} // Initialize message list for this source.

				// Check if all markers have been received from each inbound link.
				if len(server.receivedMarkers[msg.snapshotId]) == len(server.inboundLinks) {
					// If markers from all inbound links are received, conclude snapshot.
					server.inSnapshot[msg.snapshotId] = false
					delete(server.snapshotIdMap, msg.snapshotId)
					server.sim.NotifySnapshotComplete(server.Id, msg.snapshotId)
				}
			}

		}
}


// Start the chandy-lamport snapshot algorithm on this server.
// This should be called only once per server.
func (server *Server) StartSnapshot(snapshotId int) {

	if server.inSnapshot[snapshotId] {
		return // Prevent starting a new snapshot if one is already in progress.
	}

	// Record the local state and mark the snapshot as started.
	server.inSnapshot[snapshotId] = true
	server.snapshotIdMap[snapshotId] = 1
	server.receivedMarkers[snapshotId] = make(map[string]bool)
	server.stateMap[snapshotId] = server.Tokens
	server.recordedMessages[snapshotId] = make(map[string][]interface{}) // Record an empty slice for self.

	// Send a marker message to all outbound links.
	marker := MarkerMessage{snapshotId}
	server.SendToNeighbors(marker)
	fmt.Printf("Server %s sending to neighbors\n", server.Id)

}
