# Start from the latest golang base image
FROM golang:latest
# Install protoc
RUN apt-get update && apt-get install -y protobuf-compiler

# Install the protoc-gen-go and protoc-gen-go-grpc plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Generate Go code from the protobuf file in the crypto directory
RUN protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative protos/crypto.proto

# Build the Go app
RUN go build -o main .

EXPOSE 50052

# Command to run the executable
CMD ["./main"]