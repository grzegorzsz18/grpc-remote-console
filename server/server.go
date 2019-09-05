package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
)
import pb ".."

const(
	port = ":9999"
)


func main() {

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCommandServiceServer(s, &CommandService{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
