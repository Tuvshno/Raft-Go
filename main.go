package main

import (
	"flag"
	"log"
	"net"

	"github.com/Tuvshno/Raft-Go/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	id := flag.Int64("id", 1, "id of the server")
	flag.Parse()

	s := grpc.NewServer()
	reflection.Register(s)

	server := NewServer(*id)

	listener, err := net.Listen("tcp", server.cluster[*id])
	if err != nil {
		log.Fatalf("failed to listen : %s", err)
	}

	// Check if a leader exists in cluster

	go server.electionTimer()

	log.Printf("Raft Server %d listening at %s", server.id, listener.Addr())
	pb.RegisterRaftServer(s, server)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve : %s", err)
	}
}
