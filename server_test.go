package main

import (
	"context"
	"log"
	"testing"

	"github.com/Tuvshno/Raft-Go/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestAppendEntries(t *testing.T) {
	conn, err := grpc.NewClient(":8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewRaftClient(conn)
	ctx := context.Background()

	req := &pb.AppendRequest{
		Term:         1,
		LeaderId:     1,
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries:      []string{"entry1"},
		LeaderCommit: 0,
	}

	resp, err := client.AppendEntries(ctx, req)
	if err != nil {
		t.Fatalf("AppendEntries failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success to be true, got %v", resp.Success)
	}

	log.Printf("AppendEntries response: %+v", resp)
}

func TestRequestVote(t *testing.T) {
	conn, err := grpc.NewClient(":8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("did not connect : %v", err)
	}
	defer conn.Close()

	client := pb.NewRaftClient(conn)
	ctx := context.Background()

	req := &pb.VoteRequest{
		Term:         1,
		CandidateId:  1,
		LastLogIndex: 1,
		LastLogTerm:  1,
	}

	resp, err := client.RequestVote(ctx, req)
	if err != nil {
		t.Fatalf("RequestVote failed: %v", err)
	}

	if !resp.VoteGranted {
		t.Errorf("Expected voteGranted to be true, got %v", resp.VoteGranted)
	}

	log.Printf("RequestVote response: %+v", resp)
}
