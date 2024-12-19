# # Use the official Go image as a base image
# FROM golang:latest

# # Set the working directory inside the container
# WORKDIR /app

# # Copy wait-for-it.sh
# COPY wait-for-it.sh /wait-for-it.sh
# RUN chmod +x /wait-for-it.sh

# # RUN go install github.com/cosmtrek/air@latest

# # Copy go.mod and go.sum files
# COPY go.mod go.sum ./

# # Download dependencies
# RUN go mod download

# # Install grpcui inside the Docker container
# # Install grpcui inside the Docker container
# # Install grpcui using Go
# RUN GO111MODULE=on go install github.com/fullstorydev/grpcui/cmd/grpcui@latest



# # Copy the entire project
# COPY . .

# # Build the Go application
# RUN go build -o main ./cmd/oms-api/main.go

# # Expose port 8080 (or any other port your app runs on)
# EXPOSE 8080

# # Command to run the application
# # CMD ["./main"]
# CMD ["/wait-for-it.sh", "postgres-container-40:5432", "--", "./main"]




# Use the official Go image as a base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy wait-for-it.sh
COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install grpcui using Go
RUN GO111MODULE=on go install github.com/fullstorydev/grpcui/cmd/grpcui@latest

# Verify grpcui installation
RUN which grpcui
RUN grpcui --version

# Copy the entire project
COPY . .

# Build the Go application
RUN go build -o main ./cmd/oms-api/main.go

# Expose port 8089 for gRPC and 8080 for other services
EXPOSE 8089 8080

# Command to run the application
CMD ["/wait-for-it.sh", "postgres-container-40:5432", "--", "./main"]
