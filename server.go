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

type ServerOpts struct {
	ListenAddr string
}

type Server struct {
	id      int
	cluster map[int]string

	pb.UnimplementedRaftServer

	mu                sync.Mutex
	state             State
	electionTimeout   time.Duration
	lastHeartbeatTime time.Time

	currentTerm int
	votedFor    int
	log         []string

	commitIndex int
	lastIndex   int

	nextIndex  []int
	matchIndex []int
}

func NewServer(id int) *Server {
	// Hard Coded Cluster
	cluster := map[int]string{
		1: ":8001",
		2: ":8002",
		3: ":8003",
		4: ":8004",
		5: ":8005",
	}
	delete(cluster, id)

	s := &Server{
		id:      id,
		cluster: cluster,

		state: Follower,

		currentTerm: 0,
		votedFor:    -1,
		log:         make([]string, 0),

		commitIndex: -1,
		lastIndex:   -1,

		nextIndex:  make([]int, 0),
		matchIndex: make([]int, 0),
	}
	s.resetElectionTimeout()

	return s
}

func (s *Server) electionTimer() {
	for {
		s.mu.Lock()
		if time.Since(s.lastHeartbeatTime) >= s.electionTimeout {
			log.Printf("Election Timeout: %s", s.electionTimeout)
			s.state = Candidate
			log.Printf("Becoming: %s", s.state)
			s.currentTerm += 1
			log.Printf("Starting new election %d...", s.currentTerm)

			electionSuccess := false
			numVotes := 0
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
			} else {
				log.Println("Election Failed")
				s.resetElectionTimeout()
				s.state = Follower
			}
		}
		s.mu.Unlock()
	}
}

func (s *Server) resetElectionTimeout() {
	s.electionTimeout = time.Duration(rand.Intn(150)+150) * time.Millisecond
	s.lastHeartbeatTime = time.Now()
}

func (s *Server) RequestVoteToServer(ctx context.Context, port string) (bool, error) {
	conn, err := grpc.NewClient(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return false, fmt.Errorf("did not connect to %s : %v", port, err)
	}
	defer conn.Close()

	client := pb.NewRaftClient(conn)

	req := &pb.VoteRequest{
		Term:         1,
		CandidateId:  1,
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
	return &pb.AppendResponse{
		Term:    1,
		Success: true,
	}, nil
}

func (s *Server) RequestVote(ctx context.Context, in *pb.VoteRequest) (*pb.VoteResponse, error) {
	return &pb.VoteResponse{
		Term:        1,
		VoteGranted: true,
	}, nil
}
