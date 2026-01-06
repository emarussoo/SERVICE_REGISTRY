package utils

import (
	pbReg "SERVICE_REGISTRY/registry/pb"
	"context"
	"log"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func RegisterToRegistry(registryAddress string, myName string, myIp string, myPort string) (success bool) {

	conn, err := grpc.Dial(registryAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Unable to connect to registry: %v", err)
		return false
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
		return false
	} else {
		log.Printf("Service correctly registered on Registry (%s)", registryAddress)
		return true
	}

}

func UnRegisterToRegistry(registryAddress string, myName string, myIp string, myPort string) (success bool) {

	conn, err := grpc.Dial(registryAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Unable to connect to registry: %v", err)
		return false
	}
	defer conn.Close()
	client := pbReg.NewRegistryClient(conn)
	req := &pbReg.Service{
		Name:   myName,
		Ip:     myIp,
		Port:   myPort,
		Weight: 0,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = client.DropService(ctx, req)
	if err != nil {
		log.Printf("Error while registration: %v", err)
		return false
	} else {
		log.Printf("Service correctly removed from Registry (%s)", registryAddress)
		return true
	}

}
