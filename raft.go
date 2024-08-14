package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/Tuvshno/Raft-Go/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RaftState are the states that the Raft Node could be in.
type RaftState uint8

const (
	// Follower is the default state of the Raft Node.
	// Responsible for keeping a heartbeat and replicating entries from the leader.
	Follower RaftState = iota

	// Candidate is the election state of the Raft Node.
	// Responsible for starting an election and electing a new leader.
	Candidate

	// Leader is the leadership state of the Raft Node.
	// Responsible for recieving operations from the client and
	// replicating them to the Follower nodes.
	Leader
)

// String is the string representation of the RaftState.
func (raftState RaftState) String() string {
	switch raftState {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return fmt.Sprintf("%d", raftState)
	}
}

var (
	appendEntriesLogRateLimiter = 300
	appendEntriesLogCounter     = 0
)

// Raft implements a Raft Node.
type Raft struct {
	// id identifies the current Raft Node using int64 id.
	id uint64

	// cluster holds the the cluster of other raft nodes.
	cluster map[uint64]string

	// RPC Server for the Raft Node to communicate with other raft nodes.
	pb.UnimplementedRaftServer

	// applyChan is used to send logs to the main thread to be commited to the state machine
	applyCh chan *Log

	mu                sync.Mutex
	state             RaftState
	electionTimeout   time.Duration
	lastHeartbeatTime time.Time

	stateChangeCh chan RaftState
	shutdownCh    chan struct{}
	heartbeatCh   chan struct{}

	currentTerm uint64
	votedFor    uint64
	log         []string

	commitIndex int64
	lastIndex   int64

	nextIndex  []uint64
	matchIndex []uint64
}

func NewRaft(id uint64) *Raft {
	// Hard Coded Cluster for testing
	cluster := map[uint64]string{
		1: ":8001",
		2: ":8002",
		3: ":8003",
		4: ":8004",
		5: ":8005",
	}

	s := &Raft{
		id:      id,
		cluster: cluster,

		state: Follower,

		stateChangeCh: make(chan RaftState),
		shutdownCh:    make(chan struct{}),

		currentTerm: 1,
		votedFor:    0,
		log:         make([]string, 0),

		commitIndex: -1,
		lastIndex:   -1,

		nextIndex:  make([]uint64, 0),
		matchIndex: make([]uint64, 0),
	}
	s.resetElectionTimeout()

	return s
}

func (s *Raft) run() {
	for {
		select {
		case <-time.After(time.Duration(rand.Intn(150)+150) * time.Millisecond):
			s.stateChangeCh <- Candidate
			s.startElection()
		case <-s.heartbeatCh:

		case newState := <-s.stateChangeCh:
			s.handleStateChange(newState)
		case <-s.shutdownCh:
			return
		}
	}
}

func (s *Raft) handleStateChange(newState RaftState) {
	log.Printf("Server %d transition from %s to %s", s.id, s.state, newState)
	s.state = newState
}

func (s *Raft) startElection() {
	electionSuccess := false
	numVotes := 1

	for _, port := range s.cluster {
		ctx := context.Background()
		success, err := s.RequestVoteToServer(ctx, port)
		if err != nil {
			log.Println(err)
		}
		if success {
			numVotes += 1
		}
	}

	if numVotes > len(s.cluster)/2 {
		electionSuccess = true
	}

	if electionSuccess {
		s.state = Leader
		log.Println("Promoted to leader")
		go s.Lead()
	} else {
		log.Println("Election Failed")
		s.resetElectionTimeout()
		s.state = Follower
	}

}

func (s *Raft) Lead() {
	if s.state == Leader {
		for {
			for _, port := range s.cluster {
				if port == s.cluster[s.id] {
					continue
				}
				ctx := context.Background()
				go s.RequestAppendEntries(ctx, port)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (s *Raft) resetElectionTimeout() {
	s.electionTimeout = time.Duration(rand.Intn(150)+150) * time.Millisecond
	s.lastHeartbeatTime = time.Now()
}

func (s *Raft) RequestAppendEntries(ctx context.Context, port string) (bool, error) {
	conn, err := grpc.NewClient(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return false, fmt.Errorf("did not connect to %s : %v", port, err)
	}
	defer conn.Close()

	client := pb.NewRaftClient(conn)

	req := &pb.AppendRequest{
		Term:         s.currentTerm,
		LeaderId:     s.id,
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries:      nil,
		LeaderCommit: s.commitIndex,
	}

	resp, err := client.AppendEntries(ctx, req)
	if err != nil {
		return false, fmt.Errorf("AppendEntries failed to %s: %v", port, err)
	}

	// log.Printf("AppendEntries response from %s : %+v", port, resp)
	return resp.Success, nil
}

func (s *Raft) RequestVoteToServer(ctx context.Context, port string) (bool, error) {
	conn, err := grpc.NewClient(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return false, fmt.Errorf("did not connect to %s : %v", port, err)
	}
	defer conn.Close()

	client := pb.NewRaftClient(conn)

	req := &pb.VoteRequest{
		Term:         s.currentTerm,
		CandidateId:  s.id,
		LastLogIndex: 1,
		LastLogTerm:  1,
	}

	resp, err := client.RequestVote(ctx, req)
	if err != nil {
		return false, fmt.Errorf("RequestVote failed to %s: %v", port, err)
	}

	appendEntriesLogCounter++
	if appendEntriesLogCounter%appendEntriesLogRateLimiter == 0 {
		log.Printf("AppendEntries response from %s : %+v", port, resp)
	}

	log.Printf("RequestVote response: %+v", resp)
	return resp.VoteGranted, nil
}

func (s *Raft) AppendEntries(ctx context.Context, in *pb.AppendRequest) (*pb.AppendResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.resetElectionTimeout()
	if in.Term < s.currentTerm {
		return &pb.AppendResponse{
			Term:    s.currentTerm,
			Success: false,
		}, nil
	}

	s.currentTerm = in.Term
	s.state = Follower
	s.votedFor = in.LeaderId

	appendEntriesLogCounter++
	if appendEntriesLogCounter%appendEntriesLogRateLimiter == 0 {
		log.Printf("Following Leader - %d at %s", in.LeaderId, s.cluster[in.LeaderId])
	}
	return &pb.AppendResponse{
		Term:    s.currentTerm,
		Success: true,
	}, nil

}

func (s *Raft) RequestVote(ctx context.Context, in *pb.VoteRequest) (*pb.VoteResponse, error) {
	if s.currentTerm > in.Term {
		return &pb.VoteResponse{
			Term:        s.currentTerm,
			VoteGranted: false,
		}, nil
	}

	s.votedFor = in.CandidateId
	return &pb.VoteResponse{
		Term:        s.currentTerm,
		VoteGranted: true,
	}, nil
}
