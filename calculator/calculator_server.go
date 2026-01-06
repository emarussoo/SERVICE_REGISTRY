package main

import (
	pbCalc "SERVICE_REGISTRY/calculator/pb"
	"SERVICE_REGISTRY/utils"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pbCalc.CalculatorServer
}

func (s *server) Sum(ctx context.Context, req *pbCalc.Parameters) (*pbCalc.Result, error) {
	result := req.Num1 + req.Num2

	log.Printf("[gRPC] Request received: %f + %f -> Response: %f", req.Num1, req.Num2, result)

	return &pbCalc.Result{Res: result}, nil
}

func (s *server) Mul(ctx context.Context, req *pbCalc.Parameters) (*pbCalc.Result, error) {
	result := req.Num1 * req.Num2

	log.Printf("[gRPC] Request received: %f * %f -> Response: %f", req.Num1, req.Num2, result)

	return &pbCalc.Result{Res: result}, nil
}

func (s *server) Sub(ctx context.Context, req *pbCalc.Parameters) (*pbCalc.Result, error) {
	result := req.Num1 + req.Num2

	log.Printf("[gRPC] Request received: %f - %f -> Response: %f", req.Num1, req.Num2, result)

	return &pbCalc.Result{Res: result}, nil
}

func (s *server) Div(ctx context.Context, req *pbCalc.Parameters) (*pbCalc.Result, error) {
	if req.Num2 == 0 {
		return &pbCalc.Result{}, status.Errorf(codes.InvalidArgument, "Division by zero")
	}
	result := req.Num1 / req.Num2

	log.Printf("[gRPC] Request received: %f / %f -> Response: %f", req.Num1, req.Num2, result)

	return &pbCalc.Result{Res: result}, nil
}

func main() {
	ip := utils.GetOutboundIP()
	port := flag.String("port", "0", "Server port (use '0' for auto)")
	registryAddr := flag.String("registry", "localhost:50055", "Registry address ip:port")
	flag.Parse()

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("Impossible to listen on port: %v", err)
	}
	address := lis.Addr().(*net.TCPAddr)
	realPortString := strconv.Itoa(address.Port)

	s := grpc.NewServer()
	pbCalc.RegisterCalculatorServer(s, &server{})

	fmt.Printf("----------------------------------------\n")
	fmt.Printf("    SERVICE calculator IS ACTIVE\n")
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
				if utils.UnRegisterToRegistry(*registryAddr, "calculator", ip, realPortString) {
					fmt.Println(" Service removed from Registry (Invisible on registry)...")
				} else {
					fmt.Println(" An error occurred while trying to unregister from Registry...")
				}

			case "r":
				if utils.RegisterToRegistry(*registryAddr, "calculator", ip, realPortString) {
					fmt.Println(" Service added to Registry...")
				} else {
					fmt.Println(" An error occurred while trying to register to Registry...")
				}

			case "q":
				fmt.Println("Closing server...")
				utils.UnRegisterToRegistry(*registryAddr, "calculator", ip, realPortString)
				s.Stop()
				return

			default:
				fmt.Println("Invalid command")
			}
		}
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
