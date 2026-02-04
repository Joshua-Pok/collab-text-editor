#--BUILD--

FROM golang:1.25.6-alpine3.23 AS builder

WORKDIR /app

# Download dependencies

COPY go.mod go.sum ./
RUN go mod download


# Copy the rest
COPY . .

#Build binary

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app

#--RUN--

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]
