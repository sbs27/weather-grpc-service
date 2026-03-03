# --- STAGE 1: The Builder ---
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# ✅ THE FIX: Allow the 1.24 container to download 1.25 automatically
ENV GOTOOLCHAIN=auto

# Copy dependency files
COPY go.mod go.sum ./

# This will now download the Go 1.25 toolchain if go.mod requires it
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o weather-app main.go

# --- STAGE 2: The Final Runner ---
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/weather-app .
EXPOSE 50051
CMD ["./weather-app"]