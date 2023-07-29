package main

import (
	"context"

	"github.com/Tuvshno/Raft-Go/pb"
)

type State int

const (
	Leader    State = 0
	Follower  State = 1
	Candidate State = 2
)

type ServerOpts struct {
	ListenAddr string
}

type Server struct {
	// ServerOpts
	pb.UnimplementedRaftServer

	// currentTerm int
	// votedFor    int
	// log         []string

	// commitIndex int
	// lastIndex   int

	// nextIndex  []int
	// matchIndex []int
}

func NewServer() *Server {
	return &Server{}
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

// func (s *Server) Start() error {
// 	listener, err := net.Listen("tcp", s.ListenAddr)
// 	if err != nil {
// 		return fmt.Errorf("listen error: %s", err)
// 	}

// 	grpcServer := grpc.NewServer()
// 	reflection.Register(grpcServer)

// 	pb.RegisterRaftServer(grpcServer, &Server{})
// 	if err := grpcServer.Serve(listener); err != nil {
// 		return fmt.Errorf("failed to serve %s", err)
// 	}

// 	log.Printf("Server started on : %s", s.ListenAddr)
// 	return nil
// 	// for {
// 	// 	conn, err := listener.Accept()
// 	// 	if err != nil {
// 	// 		log.Printf("accept error : %s", err)
// 	// 		continue
// 	// 	}

// 	// 	go s.handleConn(conn)
// 	// }
// }

// func (s *Server) handleConn(conn net.Conn) {
// 	defer conn.Close()

// 	log.Printf("Connection made - %s <- %s", s.ListenAddr, conn.LocalAddr())

// 	buf := make([]byte, 1024)
// 	for {
// 		n, err := conn.Read(buf)
// 		if err != nil {
// 			log.Printf("read error : %s", err)
// 			break
// 		}

// 		msg := buf[:n]
// 		log.Println(string(msg))
// 	}

// }
