package main

import (
	"backend/pkg/database"
	"backend/services/user/server"
	"flag"
	"log"
	"net"

	__ "backend/proto/user"
	"google.golang.org/grpc"
)

func main() {
	database.InitMySQL()
	flag.Parse()
	lis, err := net.Listen("tcp", "127.0.0.1:9001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	__.RegisterUserServer(s, &server.UserServer{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
