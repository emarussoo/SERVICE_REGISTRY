Distributed Service Registry System using gRPC and Go.
This project implements a distributed system architecture based on the Service Registry pattern.

---> System Architecture
The project is composed of four main actors:

Service Registry: The central authority. It maintains an in-memory database of all available service instances (Name, IP, Port, Weight). It handles service registration, de-registration, and listing.

Calculator Service: A microservice that performs mathematical operations. Upon startup, it automatically finds its own IP and starts a go routing that handles basic test operations (register, deregister and close).

Echo Service: A microservice that echoes back received messages. Same logic as calculator about side operations.

Client: The entry point for the user. It contacts the Registry to discover available services and establishes a direct gRPC connection with the target service to execute requests.

---> Protocols and implementation
gRPC & Protobuf.

Dynamic Discovery: Services are not hardcoded in the client. The client asks the Registry where a service is located at runtime.

Thread-Safety: The Registry utilizes sync.Mutex to handle concurrent read/write operations on the service map safely, preventing race conditions.

Interactive Fault Injection: Service nodes include a CLI interface running in a separate Goroutine to simulate faults:

u: Unregister (Simulate failure/offline).

r: Register (Recover/Back online).

q: Graceful Shutdown.

Stateless and stateful client side load balancing (random & weighted with dummy random weights who simulate an ipothetic workload)

📂 Project Structure
Plaintext<br>
/SERVICE_REGISTRY<br>
│
├── /registry           # The central Registry Server
│   ├── /pb             # Generated gRPC code for Registry
│   └── registry.go        # Server logic
│
├── /calculator         # The Calculator Microservice
│   ├── /pb             # Generated gRPC code for Calculator
│   └── calculator_server.go        # Server logic
│
├── /echo               # The Echo Microservice
│   ├── /pb             # Generated gRPC code for Echo
│   └── echo_server.go         # Server logic
│
├── /client             # The User Client
│   └── client.go       # Logic to query Registry and call services
│
├── /utils             # Shared utilities (IP lookup, CLI listener)
│   └── utils.go
│                
└── go.mod             # Go module definition



🛠 Prerequisites
Go (1.25.4) or higher

Protoc Compiler (for regenerating .proto files, if needed)

⚡️ How to Run
To run the full system, you need to open multiple terminal windows.
The terminals can be opened in different PCs, but they need to be connected to the same LAN.

1. Start the Registry

Bash
go run registry/registry.go 

2. Start many Services as you want and use the flag -registry=<registry_ip>:<registry_port>

Bash
go run calculator/calculator_server.go -registry=<registry_ip>:<registry_port>

You can run multiple instances of the calculator or echo services on different terminals. The system will automatically assign free ports.

3. Start the client and it will automatically show the list of registered services, use same flags as before

go run client/client.go -registry=<registry_ip>:<registry_port>


---> Interactive commands
When a service (Calculator/Echo) is running, you can interact with it via its terminal:

Press u and Enter to Unregister the service from the registry (The client will no longer find it).

Press r and Enter to Register it again.

Press q and Enter to stop the server cleanly.
