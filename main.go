package main

import (
	"log"
	"net"

	"github.com/Tuvshno/Raft-Go/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("failed to listen : %s", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	server := NewServer()

	log.Printf("Raft Server listening at %s", listener.Addr())
	pb.RegisterRaftServer(s, server)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve : %s", err)
	}
}
