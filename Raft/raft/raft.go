package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// create a new Raft server.
//		rf = Make(...)
// start agreement on a new log entry
//		rf.Start(command interface{}) (index, term, isleader)
// ask a Raft for its current term, and whether it thinks it is leader
//		rf.GetState() (term, isLeader)
// each time a new entry is committed to the log, each Raft peer should send
// an ApplyMsg to the service (or tester) in the same server.
//		ApplyMsg
//

import (
	"CS345Proj3/raft/labrpc"
	"CS345Proj3/raft/labgob"
	"sync"
	"time"
	"math/rand"
	"bytes"
	// "fmt"
)
// import "bytes"
// import "labgob"

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

type AppendEntriesArguments struct {
	Term     int
	LeaderId int
	PrevIndex int
	PrevTerm  int
	LeaderCommit int
	Entries      []Log
}

type AppendEntriesReply struct {
	Term    int
	Success bool
	IndexConf int
	TermConf  int
}

type Log struct {
	Term    int
	Command interface{}
}

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]

	applyCh chan ApplyMsg         // Channel for the commit to the state machine


	// Your data here (3, 4).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	currentTerm int
	serverState int
	votedFor     int
	electionTimeout *time.Timer
	heartbeatTimeout *time.Timer

	log       []Log
	commit int
	applied int

	next  []int
	match []int


}

//
// return currentTerm and whether this server
// believes it is the leader.
//
func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	var term int
	var isleader bool

	// Your code here (3).
	term = rf.currentTerm
	isleader = rf.serverState == 2

	return term, isleader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	if err := e.Encode(rf.currentTerm); err != nil {
		panic("failed to encode currentTerm")
	}
	if err := e.Encode(rf.votedFor); err != nil {
		panic("failed to encode votedFor")
	}
	if err := e.Encode(rf.log); err != nil {
		panic("failed to encode log")
	}
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
}


func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 {
		return
	}

	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	if err := d.Decode(&rf.currentTerm); err != nil {
		panic("failed to decode currentTerm")
	}
	if err := d.Decode(&rf.votedFor); err != nil {
		panic("failed to decode votedFor")
	}
	if err := d.Decode(&rf.log); err != nil {
		panic("failed to decode log")
	}
}


//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	Term int
	CandidateID int
	LastTerm int
	LastIndex int
	// Your data here (3, 4).
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	Term int
	DidVote bool
	// Your data here (3).
}

//
// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	
	
	if args.Term < rf.currentTerm {
		reply.DidVote = false
		reply.Term = rf.currentTerm
		rf.mu.Unlock()
		return
	}

	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.votedFor = -1 // Reset votedFor
		rf.serverState = 0 // Become a follower
		rf.electionTimeout.Reset(time.Duration(300+rand.Intn(300)) * time.Millisecond) // Reset the election timeout
		rf.persist()
	}

	

	reply.Term = rf.currentTerm

	// Check if candidate's log is up to date
	lastLogIndex := len(rf.log) - 1
	lastLogTerm := rf.log[len(rf.log)-1].Term

	upToDate := args.LastTerm > lastLogTerm || (args.LastTerm == lastLogTerm && args.LastIndex >= lastLogIndex)

	if (rf.votedFor == -1 || rf.votedFor == args.CandidateID) && upToDate {
		reply.DidVote = true
		rf.votedFor = args.CandidateID

		
	} else {
		reply.DidVote = false
	}
	rf.persist()
	rf.mu.Unlock()
}



//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.serverState != 2 {
		return -1, rf.currentTerm, false
	}

	term := rf.currentTerm
	rf.log = append(rf.log, Log{term, command})
	rf.persist()
	// go rf.handleLeader()

	return len(rf.log) - 1, term, true
}


//
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (3, 4).
	rf.currentTerm = 0
	rf.serverState = 0
	rf.electionTimeout = time.NewTimer(time.Duration(300+rand.Intn(300)) * time.Millisecond)


	//Part 4
	rf.commit = 0
	rf.applied = 0
	rf.log = append(rf.log, Log{Term: 0})
	rf.applyCh = applyCh
	
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())
	rf.heartbeatTimeout = time.NewTimer(100 * time.Millisecond)
	go rf.handleServerStatus()

	return rf
}

func (rf *Raft) AppendEntries(args *AppendEntriesArguments, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	rf.electionTimeout.Reset(time.Duration(300 + rand.Intn(300)) * time.Millisecond)

	reply.Term = rf.currentTerm
	reply.Success = false
	reply.IndexConf = -1
	reply.TermConf = -1

	if args.Term < rf.currentTerm {
		return
	}
	
	
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.votedFor = -1
		rf.serverState = 0
		rf.persist()
	} else if len(args.Entries) > 0 || args.LeaderCommit > rf.commit {
		rf.persist()
	}

	
	lastIndex := len(rf.log) - 1
	if args.PrevIndex > lastIndex {
		reply.IndexConf = lastIndex + 1
		return
	}

	if args.PrevTerm != rf.log[args.PrevIndex].Term {
		for i := args.PrevIndex; i >= 0 && rf.log[i].Term == rf.log[args.PrevIndex].Term; i-- {
			reply.IndexConf = i
		}
		reply.TermConf = rf.log[args.PrevIndex].Term
		rf.log = rf.log[:args.PrevIndex] 
		rf.persist()
		return
	}


	if len(args.Entries) > 0 {
		rf.log = append(rf.log[:args.PrevIndex+1], args.Entries...)
		rf.persist()
	}
	reply.Success = true
	

	if args.LeaderCommit > rf.commit {
		if args.LeaderCommit < len(rf.log)-1 {
			rf.commit = args.LeaderCommit
		} else {
			rf.commit = len(rf.log) - 1
		}
		
		rf.persist() 
		go rf.Apply()
	}
	rf.persist()
	return
}

func (rf *Raft) handleLeader() {
    rf.mu.Lock()
    rf.electionTimeout.Stop()
    rf.electionTimeout.Reset(time.Duration(300 + rand.Intn(300)) * time.Millisecond)
    currentTerm := rf.currentTerm
    rf.mu.Unlock()

    for i := range rf.peers {
        if i != rf.me {
            go func(server int) {
                rf.mu.Lock()
             
                if rf.next[server] > len(rf.log) {
                    rf.next[server] = len(rf.log)
                }
                if rf.next[server] < 0 { 
                    rf.next[server] = 0
                }
                entries := make([]Log, len(rf.log[rf.next[server]:]))
                copy(entries, rf.log[rf.next[server]:])
                PrevIndex := rf.next[server] - 1
              
                if PrevIndex < 0 {
                    PrevIndex = 0
                }
                PrevTerm := rf.log[PrevIndex].Term
                leaderCommit := rf.commit
                rf.mu.Unlock()

                args := AppendEntriesArguments{
                    Term:         currentTerm,
                    LeaderId:     rf.me,
                    PrevIndex:    PrevIndex,
                    PrevTerm:     PrevTerm,
                    LeaderCommit: leaderCommit,
                    Entries:      entries,
                }

                var reply AppendEntriesReply
                if rf.peers[server].Call("Raft.AppendEntries", &args, &reply) {
                    rf.mu.Lock()

                    if rf.serverState != 2 {
                        rf.mu.Unlock()
                        return
                    }

                    if args.Term != currentTerm {
                        rf.mu.Unlock()
                        return
                    }

                    if reply.Term < currentTerm {
                        rf.mu.Unlock()
                        return
                    }

                    if reply.Term > currentTerm {
                        rf.currentTerm = reply.Term
                        rf.serverState = 0
                        rf.votedFor = -1
                        rf.persist()
                        rf.mu.Unlock()
                        return
                    }

                    if reply.Success {
                        newMatchIndex := args.PrevIndex + len(args.Entries)
                        if newMatchIndex > rf.match[server] {
                            rf.match[server] = newMatchIndex
                            rf.next[server] = rf.match[server] + 1
                 
                        }
                    } else if reply.TermConf < 0 {
                        rf.next[server] = reply.IndexConf
                        rf.match[server] = rf.next[server] - 1
                    
                    } else {
                        newNextIndex := len(rf.log) - 1
                        for ; newNextIndex >= 0; newNextIndex-- {
                            if rf.log[newNextIndex].Term == reply.TermConf {
                                break
                            }
                        }
                        if newNextIndex < 0 {
                            rf.next[server] = reply.IndexConf
                        } else {
                            rf.next[server] = newNextIndex + 1
                        }
                        rf.match[server] = rf.next[server] - 1
                    
                    }

                    for n := len(rf.log) - 1; n >= rf.commit; n-- {
                        count := 1
                        for i := 0; i < len(rf.peers); i++ {
                            if i != rf.me && rf.match[i] >= n {
                                count++
                            }
                        }
                        if count > len(rf.peers)/2 && rf.log[n].Term == rf.currentTerm {
                            rf.commit = n
                            rf.persist()  
                            go rf.Apply()
                            break
                        }
                    }

                    rf.mu.Unlock()
                }
            }(i)
        }
    }
}

func (rf *Raft) Apply() {
    rf.mu.Lock()
    defer rf.mu.Unlock()
   
    startIndex := rf.applied + 1
    endIndex := rf.commit
    lastIndex := len(rf.log) - 1
    
    if endIndex > lastIndex {
        endIndex = lastIndex
    }

    if startIndex > endIndex {
        // No new entries to apply
        return
    }

    entriesToApply := make([]ApplyMsg, 0, endIndex-startIndex+1)
    for i := startIndex; i <= endIndex; i++ {
        if i < len(rf.log) {
            entriesToApply = append(entriesToApply, ApplyMsg{
                CommandValid: true,
                Command:      rf.log[i].Command,
                CommandIndex: i,
            })
        } 
    }
    rf.applied = endIndex


    rf.mu.Unlock()
    for _, msg := range entriesToApply {
        rf.applyCh <- msg
    }
    rf.mu.Lock() 
}



func (rf *Raft) handleNotLeader() {
    rf.mu.Lock()
	defer rf.mu.Unlock()
	//fmt.Printf("%d election timeout for term %d\n", rf.me, rf.currentTerm)
	rf.electionTimeout.Reset(time.Duration(300+rand.Intn(300)) * time.Millisecond)

	if rf.serverState == 2 { // Do not start election on leader
		return
	}

	newTerm := rf.currentTerm + 1
	//fmt.Printf("%d is starting new election for term %d\n", rf.me, newTerm)
	rf.currentTerm = newTerm
	rf.votedFor = rf.me
	votes := 1 
	totalPeers := len(rf.peers)
	args := RequestVoteArgs{
		Term: newTerm, 
		CandidateID: rf.me,
		LastIndex: len(rf.log) - 1,
		LastTerm: rf.log[len(rf.log) - 1].Term}
	rf.serverState = 1 
	rf.persist()

	for peer := 0; peer < totalPeers; peer++ {
		//fmt.Printf("%d SENDING vote to %d for term %d\n", rf.me, peer, newTerm)
		if peer != rf.me {
			go func(peerId int) {
				voteReply := RequestVoteReply{}
				
				if rf.sendRequestVote(peerId, &args, &voteReply) {
					rf.mu.Lock()
                    defer rf.mu.Unlock()
					currentState := rf.serverState
					if voteReply.Term > newTerm {

						rf.currentTerm = voteReply.Term
						rf.serverState = 0
						rf.votedFor = -1
						rf.persist()
					
						return

					}
					if voteReply.DidVote && voteReply.Term == newTerm {
						votes++
						if votes > totalPeers/2 && currentState != 2 {
							//fmt.Printf("%d got elected for term %d\n", rf.me, rf.currentTerm)
							rf.serverState = 2 // Transition to leader
							
							rf.next = make([]int, len(rf.peers))
							rf.match = make([]int, len(rf.peers))
							lastIndex := len(rf.log)
							for i := range rf.peers {
								rf.next[i] = lastIndex
							}
							// go rf.handleLeader()
							return 
							
						}
					} 
				}
			}(peer)
		}
	}
}


func (rf *Raft) handleServerStatus() {
    for {
        select {
        case <-rf.heartbeatTimeout.C:
            rf.mu.Lock()
            rf.heartbeatTimeout.Reset(time.Duration(100) * time.Millisecond)
            currentState := rf.serverState
            rf.mu.Unlock()

            if currentState == 2 {
                rf.mu.Lock()
                rf.persist()
                rf.mu.Unlock()

                rf.handleLeader()
            }

            rf.mu.Lock()
            rf.heartbeatTimeout.Reset(time.Duration(100) * time.Millisecond)
            rf.mu.Unlock()

        case <-rf.electionTimeout.C:
            rf.handleNotLeader()
        }
    }
}




