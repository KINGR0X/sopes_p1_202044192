FROM rust:latest AS builder 

# Install protobuf compiler
RUN apt-get update && apt-get install -y protobuf-compiler && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/src/app

# Copy the entire project
COPY . .

# Build the application
RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim

# Install necessary runtime libraries
RUN apt-get update && apt-get install -y libssl3 ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /usr/src/app/target/release/api_rust .

# Set the startup command
CMD ["./api_rust"]