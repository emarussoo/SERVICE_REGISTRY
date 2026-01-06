package main

import (
	pbEcho "SERVICE_REGISTRY/echo/pb"
	pbReg "SERVICE_REGISTRY/registry/pb"
	"SERVICE_REGISTRY/utils"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	pbEcho.UnimplementedEchoServer
}

func (s *server) Repeat(ctx context.Context, req *pbEcho.Msg) (*pbEcho.Reply, error) {
	msg := req.Mess
	log.Printf("[gRPC] Received request: Repeat: %s", msg)
	return &pbEcho.Reply{Reply: msg}, nil
}

func registerToRegistry(registryAddress string, myName string, myIp string, myPort string) {

	conn, err := grpc.Dial(registryAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Unable to connect to registry: %v", err)
		return
	}
	defer conn.Close()

	client := pbReg.NewRegistryClient(conn)

	req := &pbReg.Service{
		Name:   myName,
		Ip:     myIp,
		Port:   myPort,
		Weight: (int32)(rand.Intn(100)),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = client.RegisterService(ctx, req)
	if err != nil {
		log.Printf("Error during registration: %v", err)
	} else {
		log.Printf("Service registered on registry (%s)", registryAddress)
	}
}

func main() {

	ip := utils.GetOutboundIP()
	port := flag.String("port", "0", "Server port (use '0' for auto)")

	registryAddr := flag.String("registry", "localhost:50055", "Registry address ip:port")
	flag.Parse()

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("IUnable to listen on port: %v", err)
	}
	address := lis.Addr().(*net.TCPAddr)
	realPortInt := address.Port
	realPortString := strconv.Itoa(realPortInt)

	s := grpc.NewServer()
	pbEcho.RegisterEchoServer(s, &server{})

	fmt.Printf("----------------------------------------\n")
	fmt.Printf("    SERVICE echo IS ACTIVE\n")
	fmt.Printf("    Ip: %s\n", ip)
	fmt.Printf("    Port:     %s\n", realPortString)
	fmt.Printf("    Protocol: gRPC\n")
	fmt.Printf("----------------------------------------\n")

	go func() {
		var input string
		for {
			fmt.Println("\n--- COMMANDS ---")
			fmt.Println("[u] Unregister (Invisible on regisrty)")
			fmt.Println("[r] Register   (Visible on registry)")
			fmt.Println("[q] Quit       (Close all)")
			fmt.Print("Command: ")

			fmt.Scan(&input)

			switch input {
			case "u":
				if utils.UnRegisterToRegistry(*registryAddr, "echo", ip, realPortString) {
					fmt.Println(" Service removed from Registry (Invisible on registry)...")
				} else {
					fmt.Println(" An error occured while trying to unregister from Registry...")
				}

			case "r":
				if utils.RegisterToRegistry(*registryAddr, "echo", ip, realPortString) {
					fmt.Println(" Service added to Registry...")
				} else {
					fmt.Println(" An error occured while trying to register to Registry...")
				}

			case "q":
				fmt.Println("Closing server...")
				utils.UnRegisterToRegistry(*registryAddr, "echo", ip, realPortString)
				s.Stop()
				return

			default:
				fmt.Println("Invalid command")
			}
		}
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Errore server: %v", err)
	}
}
