# Use the official Golang base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files to the container
COPY go.mod .
COPY go.sum .

# Download dependencies
RUN go mod download

# Copy the entire project source code to the container
COPY . .

# Build the forum application
RUN go build -o myforum 

# Expose the port on which the forum application runs
EXPOSE 8080

# Run the forum application
CMD ["./myforum"]
