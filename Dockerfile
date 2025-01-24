# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.23.4 AS build-stage

WORKDIR /app

# COPY go.mod go.sum ./

COPY . .

RUN go mod download

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/health_check

# Final stage
FROM gcr.io/distroless/base-debian11

# Set the working directory
WORKDIR /

# Copy the binary and sample input.yaml from the build stage
COPY --from=build-stage /app/cmd/health_check/health_check .
COPY --from=build-stage /app/input.yaml .

ENTRYPOINT [ "./health_check" ]

CMD ["--config-file=./input.yaml"]

