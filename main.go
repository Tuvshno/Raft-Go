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
	id := flag.Int("id", 1, "id of the server")
	flag.Parse()

	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("failed to listen : %s", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	server := NewServer(*id)

	go server.electionTimer()

	log.Printf("Raft Server %d listening at %s", server.id, listener.Addr())
	pb.RegisterRaftServer(s, server)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve : %s", err)
	}
}
