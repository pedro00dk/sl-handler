
#build stage
FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache gcc musl-dev git
RUN go get -d -v ./...
WORKDIR /app/src
RUN GOOS=linux go build -a -o ./app -- main.go

#final stage
FROM alpine:latest
RUN apk --no-cache add gcc musl-dev ca-certificates
WORKDIR /app
COPY --from=builder . .
CMD ./app
LABEL Name=tg Version=0.0.1
EXPOSE 8000
