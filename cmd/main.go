package main

import (
	"fmt"
	"log"
	"net"

	"mobilehub-order/pkg/config"
	"mobilehub-order/pkg/db"
	pb "mobilehub-order/pkg/pb"
	services "mobilehub-order/pkg/services"

	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc"
)

func main() {
	c, err := config.LoadConfig()

	if err != nil {
		log.Fatalln("Failed at config", err)
	}

	h := db.Init(c.DBUrl)

	lis, err := net.Listen("tcp", c.Port)

	if err != nil {
		log.Fatalln("Failed to listing:", err)
	}

	fmt.Println("Order Svc on", c.Port)

	s := services.OrderServiceServer{
		H: h,
	}

	grpcServer := grpc.NewServer()

	pb.RegisterOrderServiceServer(grpcServer, &s)
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}
}
