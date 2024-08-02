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

type State string

const (
	Follower  State = "Follower"
	Candidate State = "Candidate"
	Leader    State = "Leader"
)

type Server struct {
	id      int64
	cluster map[int64]string

	pb.UnimplementedRaftServer

	mu                sync.Mutex
	state             State
	electionTimeout   time.Duration
	lastHeartbeatTime time.Time

	currentTerm int64
	votedFor    int64
	log         []string

	commitIndex int64
	lastIndex   int64

	nextIndex  []int64
	matchIndex []int64
}

func NewServer(id int64) *Server {
	// Hard Coded Cluster for testing
	cluster := map[int64]string{
		1: ":8001",
		2: ":8002",
		3: ":8003",
		4: ":8004",
		5: ":8005",
	}

	s := &Server{
		id:      id,
		cluster: cluster,

		state: Follower,

		currentTerm: 0,
		votedFor:    -1,
		log:         make([]string, 0),

		commitIndex: -1,
		lastIndex:   -1,

		nextIndex:  make([]int64, 0),
		matchIndex: make([]int64, 0),
	}
	s.resetElectionTimeout()

	return s
}

func (s *Server) electionTimer() {
	for {
		s.mu.Lock()
		if s.state != Leader && time.Since(s.lastHeartbeatTime) >= s.electionTimeout {
			s.state = Candidate
			s.currentTerm++
			log.Printf("Starting new election %d...", s.currentTerm)
			s.startElection()
		}
		s.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *Server) startElection() {
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

func (s *Server) Lead() {
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

func (s *Server) resetElectionTimeout() {
	s.electionTimeout = time.Duration(rand.Intn(150)+150) * time.Millisecond
	s.lastHeartbeatTime = time.Now()
}

func (s *Server) RequestAppendEntries(ctx context.Context, port string) (bool, error) {
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

	log.Printf("AppendEntries response: %+v", resp)
	return resp.Success, nil
}

func (s *Server) RequestVoteToServer(ctx context.Context, port string) (bool, error) {
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

	log.Printf("RequestVote response: %+v", resp)
	return resp.VoteGranted, nil
}

func (s *Server) AppendEntries(ctx context.Context, in *pb.AppendRequest) (*pb.AppendResponse, error) {
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

	return &pb.AppendResponse{
		Term:    s.currentTerm,
		Success: true,
	}, nil

}

func (s *Server) RequestVote(ctx context.Context, in *pb.VoteRequest) (*pb.VoteResponse, error) {
	if s.currentTerm > in.Term {
		return &pb.VoteResponse{
			Term:        s.currentTerm,
			VoteGranted: false,
		}, nil
	}

	return &pb.VoteResponse{
		Term:        s.currentTerm,
		VoteGranted: true,
	}, nil
}
