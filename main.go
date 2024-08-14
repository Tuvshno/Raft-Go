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

	raft := NewRaft(*id)

	listener, err := net.Listen("tcp", raft.cluster[*id])
	if err != nil {
		log.Fatalf("failed to listen : %s", err)
	}

	go raft.run()

	log.Printf("Raft Server %d listening at %s", raft.id, listener.Addr())
	pb.RegisterRaftServer(s, raft)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve : %s", err)
	}
}
