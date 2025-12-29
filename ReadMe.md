# OMS API - Order Management System

This project is a Go-based Order Management System (OMS) API using gRPC. It provides services for managing items, users, and orders with PostgreSQL as the database backend.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Project Structure](#project-structure)
- [Setup Instructions](#setup-instructions)
- [Running the Project](#running-the-project)
- [Docker Setup](#docker-setup)
- [Environment Variables](#environment-variables)
- [gRPC Setup](#grpc-setup)
- [API Access](#api-access)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go** 1.22.2 or higher ([Download](https://go.dev/dl/))
- **PostgreSQL** 12 or higher ([Download](https://www.postgresql.org/download/))
- **Protocol Buffers Compiler (protoc)** ([Download](https://grpc.io/docs/protoc-installation/))
- **Go plugins for protoc**:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```
- **Docker & Docker Compose** (optional, for containerized setup)

---

## Project Structure

```
oms-grpc/
├── cmd/oms-api/
│   ├── handlers/          # gRPC service handlers
│   │   ├── item_handler.go
│   │   ├── order_handler.go
│   │   └── user_handler.go
│   ├── models/            # Data models
│   │   ├── items.go
│   │   ├── orders.go
│   │   ├── users.go
│   │   └── discount.go
│   ├── proto/             # Protocol buffer definitions
│   │   ├── oms_items.proto
│   │   ├── oms_order.proto
│   │   └── oms_users.proto
│   ├── protobuf/          # Generated Go protobuf files
│   ├── protobufJs/       # Generated JavaScript protobuf files
│   ├── utils/            # Utility functions
│   ├── scripts/          # Helper scripts
│   ├── main.go           # Application entry point
│   └── Makefile          # Commands for generating protobuf files
├── docker-compose.yml    # Docker Compose configuration
├── Dockerfile           # Docker image definition
├── go.mod               # Go module dependencies
├── .env.example         # Environment variables template
└── ReadMe.md            # This file
```

---

## Setup Instructions

### 1. Clone the Repository

```bash
git clone <repository-url>
cd oms-grpc
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Setup PostgreSQL Database

#### Option A: Local PostgreSQL Installation

1. Install PostgreSQL on your system
2. Create a database:
   ```bash
   createdb oms
   ```
3. Or using psql:
   ```bash
   psql -U postgres
   CREATE DATABASE oms;
   ```

#### Option B: Using Docker (Recommended for Development)

PostgreSQL will be automatically set up when using Docker Compose (see [Docker Setup](#docker-setup)).

### 4. Configure Environment Variables

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` file with your database credentials:
   ```bash
   # For local development
   DB_HOST=localhost
   DB_PORT=5433
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=oms
   ```

See [Environment Variables](#environment-variables) section for all available options.

### 5. Generate Protocol Buffer Files

Navigate to the `cmd/oms-api` directory and generate protobuf files:

```bash
cd cmd/oms-api
make generate_proto_files
```

Or manually:
```bash
for proto_file in proto/*.proto; do \
    protoc --proto_path=proto \
    --go_out=paths=source_relative:./protobuf \
    --go-grpc_out=paths=source_relative:./protobuf \
    $proto_file; \
done
```

---

## Running the Project

### Local Development

#### Option 1: From cmd/oms-api directory
```bash
cd cmd/oms-api
go run main.go
```

#### Option 2: From project root
```bash
go run cmd/oms-api/main.go
```

The application will:
- Connect to PostgreSQL database
- Start gRPC server on port `8089`
- Start grpcui web interface on port `8080` (if enabled)

### Build and Run Binary

```bash
cd cmd/oms-api
go build -o oms-api main.go
./oms-api
```

---

## Docker Setup

### Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+

### Quick Start with Docker

1. **Build and start all services:**
   ```bash
   docker-compose up --build
   ```

2. **Run in detached mode (background):**
   ```bash
   docker-compose up -d --build
   ```

3. **View logs:**
   ```bash
   docker-compose logs -f oms-api
   ```

4. **Stop services:**
   ```bash
   docker-compose down
   ```

5. **Stop and remove volumes (clean slate):**
   ```bash
   docker-compose down -v
   ```

### Docker Services

The `docker-compose.yml` includes:

- **postgres-service**: PostgreSQL 15 database
  - Port: `5432`
  - Database: `oms`
  - User: `root` / Password: `root`

- **oms-api**: OMS gRPC API service
  - gRPC Server: `localhost:8089`
  - gRPC Web UI: `http://localhost:8080`

### Docker Environment Variables

Docker Compose automatically configures environment variables for containerized setup. The `DB_HOST` is set to `postgres-service` (container name) for Docker networking.

---

## Environment Variables

The application supports the following environment variables (with defaults):

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | Database host address (use `postgres-service` for Docker) |
| `DB_PORT` | `5433` | Database port (`5432` for Docker) |
| `DB_USER` | `postgres` | Database username |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `oms` | Database name |

### gRPC Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GRPC_PORT` | `8089` | gRPC server port |
| `GRPC_HOST` | `localhost` | gRPC host address (for grpcui connection) |

### gRPC UI Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ENABLE_GRPCUI` | `true` | Enable/disable grpcui web interface |
| `GRPCUI_PORT` | `8080` | grpcui web interface port |

### Using Environment Variables

**Local Development:**
- Create `.env` file from `.env.example`
- Variables are automatically loaded when running `go run`

**Docker:**
- Environment variables are set in `docker-compose.yml`
- Override by editing the `environment` section

---

## gRPC Setup

### 1. Update Go Path in `.bashrc` (if not already done)

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
```

### 2. Install Protocol Buffer Compiler

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install -y protobuf-compiler

# Verify installation
protoc --version
```

**macOS:**
```bash
brew install protobuf
```

**Windows:**
Download from [Protocol Buffers Releases](https://github.com/protocolbuffers/protobuf/releases)

### 3. Install Go Protobuf Plugins

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 4. Generate Protobuf Files

```bash
cd cmd/oms-api
make generate_proto_files
```

---

## API Access

### gRPC Server

- **Port**: `8089`
- **Protocol**: gRPC (plaintext)
- **Reflection**: Enabled (for grpcui)

### gRPC Web UI (grpcui)

Access the web interface at:
- **URL**: `http://localhost:8080`
- **Features**:
  - Browse available services and methods
  - Test gRPC calls interactively
  - View request/response data

### Available Services

1. **OmsItemService**: Item management operations
2. **UserService**: User management operations
3. **OrderService**: Order management operations

---

## GRPCUI Installation Guide

### Option 1: Install Pre-Built Binary

1. **Download the tar file**:
   - Visit the [grpcui releases page](https://github.com/fullstorydev/grpcui/releases)
   - Download the appropriate tar file (e.g., `grpcui_1.4.2_linux_x86_64.tar.gz`)

2. **Extract and install**:
   ```bash
   tar -xvf grpcui_1.4.2_linux_x86_64.tar.gz
   sudo mv grpcui /usr/local/bin/
   sudo chmod +x /usr/local/bin/grpcui
   ```

3. **Verify installation**:
   ```bash
   grpcui -version
   ```

### Option 2: Build from Source

```bash
git clone https://github.com/fullstorydev/grpcui.git
cd grpcui/cmd/grpcui
go build -o /usr/local/bin/grpcui
sudo chmod +x /usr/local/bin/grpcui
```

### Option 3: Install via Go (Recommended)

```bash
go install github.com/fullstorydev/grpcui/cmd/grpcui@latest
```

**Note**: grpcui is automatically included in the Docker image, so no manual installation is needed when using Docker.

---

## Troubleshooting

### Database Connection Issues

- **Error**: "Failed to connect to database"
  - **Solution**: Verify PostgreSQL is running and credentials in `.env` are correct
  - Check database exists: `psql -U postgres -l`
  - For Docker: Ensure `postgres-service` container is running

### Port Already in Use

- **Error**: "Failed to create listener: address already in use"
  - **Solution**: Change port in `.env` or stop the process using the port
  - Find process: `lsof -i :8089` or `netstat -tulpn | grep 8089`

### Protobuf Generation Errors

- **Error**: "protoc: command not found"
  - **Solution**: Install Protocol Buffer Compiler (see [gRPC Setup](#grpc-setup))

- **Error**: "protoc-gen-go: program not found"
  - **Solution**: Install Go protobuf plugins:
    ```bash
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    ```

### Docker Issues

- **Error**: "Cannot connect to database"
  - **Solution**: Wait for PostgreSQL to be ready (healthcheck ensures this)
  - Check logs: `docker-compose logs postgres-service`

- **Error**: "grpcui not accessible"
  - **Solution**: Ensure port `8080` is not blocked by firewall
  - Verify container is running: `docker-compose ps`

### gRPC Web UI Not Loading

- **Error**: "This site can't be reached"
  - **Solution**: 
    - For Docker: Ensure `ENABLE_GRPCUI=true` in environment
    - Check if grpcui is running: `docker-compose logs oms-api | grep grpcui`
    - Verify port mapping in `docker-compose.yml`

### General Tips

- Ensure Go binary path is in `$PATH`: `export PATH="$PATH:$(go env GOPATH)/bin"`
- Verify all dependencies: `go mod verify`
- Check Docker logs: `docker-compose logs -f`
- Restart services: `docker-compose restart`

---

## Additional Resources

- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers Guide](https://developers.google.com/protocol-buffers)
- [GORM Documentation](https://gorm.io/docs/)
- [grpcui GitHub](https://github.com/fullstorydev/grpcui)

---

## License

[Add your license information here]
