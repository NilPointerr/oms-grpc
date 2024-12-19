
# OMS API - Order Management System

This project is a Go-based Order Management System (OMS) API. Below are the instructions for running the project, generating protobuf files, and setting up gRPC.

## Project Structure
```
cmd/oms-api/
├── handlers/        # HTTP and gRPC handlers
├── models/          # Data models
├── routes/          # HTTP and gRPC route definitions
├── utils/           # Utility functions
├── grpc/            # gRPC service implementations
│   ├── order_service.go # gRPC service implementation
├── proto/           # .proto files
│   ├── oms_service.proto
├── protobuf/        # Generated proto files
│   ├── oms_service.pb.go
│   ├── oms_service_grpc.pb.go
├── main.go          # Application entry point
├── Makefile         # For generating protobuf files
```

---

## Running the Project
### Option 1
1. Navigate to the `cmd/oms-api` directory:
   ```bash
   cd cmd/oms-api
   ```
2. Run the following command to start the application:
   ```bash
   go run main.go
   ```

### Option 2
1. From the root `OMS` directory, run the following command:
   ```bash
   go run cmd/oms-api/main.go
   ```

### Using Docker
1. Build the Docker image:
   ```bash
   sudo docker build -t oms-api-g .
   ```

---

## gRPC Setup Guide for OMS Service
### 1. Update the `.bashrc` File for Go Path
1. Open the `.bashrc` file:
   ```bash
   nano ~/.bashrc
   ```
2. Add the following line to include the Go binary path:
   ```bash
   export PATH="$PATH:$(go env GOPATH)/bin"
   ```
3. Save the file and reload it:
   ```bash
   source ~/.bashrc
   ```

### 2. Generate Proto Files
#### Option 1: Using Makefile
1. Navigate to the `./oms-api` directory:
   ```bash
   cd ./oms-api
   ```
2. Run the following command to generate proto files:
   ```bash
   make generate_proto_files
   ```

#### Option 2: Using `protoc`
1. Navigate to the `./oms-api` directory:
   ```bash
   cd ./oms-api
   ```
2. Run the following command to manually generate proto files:
   ```bash
   for proto_file in proto/*.proto; do \
       protoc --proto_path=proto \
       --go_out=paths=source_relative:./protobuf \
       --go-grpc_out=paths=source_relative:./protobuf \
       $$proto_file; \
   done
   ```

### 3. Run the OMS Service
1. Navigate to the `./oms-api` directory:
   ```bash
   cd ./oms-api
   ```
2. Run the service:
   ```bash
   go run main.go
   ```

---

## GRPCUI Installation Guide for Linux
### Option 1: Install Pre-Built Binary
1. **Download the tar file**:
   - Visit the [grpcui releases page](https://github.com/fullstorydev/grpcui/releases) and download the appropriate tar file for your system (e.g., `grpcui_1.4.2_linux_x86_64.tar.gz`).

2. **Extract the tar file**:
   ```bash
   tar -xvf grpcui_1.4.2_linux_x86_64.tar.gz
   ```

3. **Move the binary to a system-wide location**:
   ```bash
   sudo mv grpcui /usr/local/bin/
   ```

4. **Make the binary executable**:
   ```bash
   sudo chmod +x /usr/local/bin/grpcui
   ```

5. **Verify the installation**:
   ```bash
   grpcui -version
   ```

### Option 2: Build grpcui from Source
1. **Clone the repository**:
   ```bash
   git clone https://github.com/fullstorydev/grpcui.git
   ```

2. **Navigate to the grpcui command directory**:
   ```bash
   cd grpcui/cmd/grpcui
   ```

3. **Build the binary**:
   ```bash
   go build -o /usr/local/bin/grpcui
   ```

4. **Make the binary executable**:
   ```bash
   sudo chmod +x /usr/local/bin/grpcui
   ```

5. **Verify the installation**:
   ```bash
   grpcui -version
   ```

---

## Usage
1. Use grpcui to connect to a gRPC server running on `localhost` (replace `12345` with the actual port):
   ```bash
   grpcui -plaintext localhost:12345
   ```

This will open a web-based UI for interacting with the gRPC server.

---

## Troubleshooting
- Ensure the `grpcui` binary is in `/usr/local/bin/`.
- Verify the binary is executable (`chmod +x`).
- Check permissions and use `sudo` if necessary.
- Confirm the gRPC server is running and accessible.

For more details, visit the official [grpcui GitHub repository](https://github.com/fullstorydev/grpcui).




air -c air.toml