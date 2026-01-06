package main

import (
	pbCalc "SERVICE_REGISTRY/calculator/pb"
	pbEcho "SERVICE_REGISTRY/echo/pb"
	"SERVICE_REGISTRY/registry/pb"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func randomElement(list []*pb.Service) *pb.Service {
	if len(list) == 0 {
		return nil
	}
	element := rand.Intn(len(list))
	return list[element]
}

func filterList(name string, list []*pb.Service) []*pb.Service {
	var filtered []*pb.Service
	for _, service := range list {
		if service.Name == name {
			filtered = append(filtered, service)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func invokeCalculator(conn *grpc.ClientConn) {

	c := pbCalc.NewCalculatorClient(conn)

	ctx := context.Background()

	res, err := c.Sum(ctx, &pbCalc.Parameters{Num1: 10, Num2: 20})

	if err != nil {
		log.Printf("RPC Calculator error: %v", err)
	} else {
		fmt.Printf("Calculator result: %f\n", res.Res)
	}
}

func invokeEcho(conn *grpc.ClientConn) {
	c := pbEcho.NewEchoClient(conn)

	ctx := context.Background()

	fmt.Println("Insert a string to repeat")
	var toRepeat string
	_, err := fmt.Scan(&toRepeat)
	if err != nil {
		fmt.Println("Scan error")
	}
	res, err := c.Repeat(ctx, &pbEcho.Msg{Mess: toRepeat})

	if err != nil {
		log.Printf("rpc call error: %v", err)
	} else {
		log.Printf("done, string %s repeated", res)
	}
}
func main() {
	registryAddr := flag.String("registry", "localhost:50055", "Registry address ip:port")
	flag.Parse()

	fmt.Printf("Trying connection to %s...\n", *registryAddr)

	conn, err := grpc.Dial(*registryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Unable to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewRegistryClient(conn)

	ctx := context.Background()

	req := &pb.Empty{}

	resp, err := client.ListServices(ctx, req)
	if err != nil {
		log.Fatalf("Call error: %v", err)
	}

	for _, service := range resp.List {
		log.Printf(" %s---> %s:%s (Peso: %d)", service.Name, service.Ip, service.Port, service.Weight)

	}

	var name string
	var choice int
	var chosenService *pb.Service
	var filteredList []*pb.Service
	for {
		fmt.Print("Enter service name: ")
		_, err = fmt.Scan(&name)
		if err != nil {
			log.Fatalf("Scan error: %v", err)
		}
		filteredList = filterList(name, resp.List)
		if len(filteredList) == 0 {
			log.Printf("No services found")
		} else {
			break
		}
	}

	fmt.Println("Select a lb method: (random by default)")
	fmt.Println("1 - Random")
	fmt.Println("2 - Weighted")

	_, err = fmt.Scan(&choice)

	if err == nil {
		switch choice {
		case 1:
			chosenService = randomElement(filteredList)
			break
		case 2:
			chosenService = filteredList[0]
			for _, service := range filteredList {
				if service.Weight < chosenService.Weight {
					chosenService = service
				}
			}
			break
		default:
			chosenService = randomElement(filteredList)
			break
		}
	}

	calcAddress := chosenService.Ip + ":" + chosenService.Port

	newConn, err := grpc.Dial(calcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Unable to connect: %v", err)
	}
	defer newConn.Close()
	switch name {
	case "calculator":
		invokeCalculator(newConn)
		break
	case "echo":
		invokeEcho(newConn)
		break
	default:
		break
	}
}
