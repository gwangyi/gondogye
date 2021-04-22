FROM golang:alpine AS builder

RUN apk add --no-cache git make build-base

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
#COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o app -ldflags="-extldflags=-static" .

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/app .

# Build a small image
FROM scratch

COPY --from=builder /dist/app /

# Command to run
EXPOSE 10000
USER 1000
ENTRYPOINT ["/app", "-port", "10000"]
