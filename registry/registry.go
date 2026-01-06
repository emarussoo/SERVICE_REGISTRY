package main

import (
	pbReg "SERVICE_REGISTRY/registry/pb"
	"SERVICE_REGISTRY/utils"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type registry_server struct {
	pbReg.UnimplementedRegistryServer
	mu       sync.Mutex
	services map[string][]*pbReg.Service
}

func (s *registry_server) RegisterService(context context.Context, req *pbReg.Service) (*pbReg.IsDone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.services == nil {
		s.services = make(map[string][]*pbReg.Service)
	}

	currentList := s.services[req.Name]

	for _, existingService := range currentList {
		if existingService.Ip == req.Ip && existingService.Port == req.Port {
			log.Printf("The service %s, %s:%s is already registered", req.Name, req.Ip, req.Port)

			return &pbReg.IsDone{Done: false}, status.Errorf(codes.AlreadyExists, "service %s already registered", req.Name)
		}
	}

	s.services[req.Name] = append(s.services[req.Name], req)
	log.Printf("\nRegistered service %s: %s, %s\n", req.Name, req.Ip, req.Port, req.Weight)

	s.showRegistryState()
	return &pbReg.IsDone{Done: true}, nil
}

func (s *registry_server) DropService(context context.Context, req *pbReg.Service) (*pbReg.IsDone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	findServiceName, exists := s.services[req.Name]

	if !exists {
		return &pbReg.IsDone{Done: false}, status.Errorf(codes.NotFound, "service %s not found", req.Name)
	}

	toReturnList := make([]*pbReg.Service, 0, len(findServiceName))

	found := false
	for _, service := range findServiceName {
		if service.Port == req.Port && service.Ip == req.Ip {
			found = true
			continue
		}
		toReturnList = append(toReturnList, service)
	}
	s.services[req.Name] = toReturnList

	var finalError error
	if !found {
		log.Printf("\nService does not exist\n")
		finalError = status.Errorf(codes.NotFound, "Service %s not found", req.Name)
	} else {
		log.Printf("\nRemoved service %s: %s, %s\n", req.Name, req.Ip, req.Port)
	}

	s.showRegistryState()
	return &pbReg.IsDone{Done: found}, finalError
}

func (s *registry_server) ListServices(context context.Context, req *pbReg.Empty) (*pbReg.ListOfServices, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var flattenedList []*pbReg.Service
	for _, instance := range s.services {
		flattenedList = append(flattenedList, instance...)
	}
	return &pbReg.ListOfServices{List: flattenedList}, nil
}

func (s *registry_server) UpdateWeight(context context.Context, req *pbReg.Service) (*pbReg.IsDone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newWeight := rand.Intn(100)
	req.Weight = int32(newWeight)
	return &pbReg.IsDone{Done: true}, nil
}

func (s *registry_server) showRegistryState() {
	log.Println("------ REGISTRY STATE ------")

	if len(s.services) == 0 {
		log.Println("   No services registered   ")
	}

	var flattenedList []*pbReg.Service
	for _, instance := range s.services {
		flattenedList = append(flattenedList, instance...)
	}

	for _, service := range flattenedList {
		log.Printf(" %s---> %s:%s (Peso: %d)", service.Name, service.Ip, service.Port, service.Weight)

	}
	log.Println("-----------------------------------------")
}

func main() {
	port := ":50055"

	ip := utils.GetOutboundIP()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Unable to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()

	registryServer := &registry_server{
		services: make(map[string][]*pbReg.Service),
	}

	pbReg.RegisterRegistryServer(grpcServer, registryServer)

	fmt.Printf("----------------------------------------\n")
	fmt.Printf("    SERVICE REGISTRY IS ACTIVE\n")
	fmt.Printf("    Ip: %s\n", ip)
	fmt.Printf("    Port:     %s\n", port)
	fmt.Printf("    Protocol: gRPC\n")
	fmt.Printf("----------------------------------------\n")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
