FROM golang:1.24 AS builder

WORKDIR /app
ADD src .
RUN go build .


FROM debian:13-slim

WORKDIR /app
COPY --from=builder /app/s2dnglobby .

RUN install -Dm755 ./s2dnglobby /bin

RUN apt-get update -y
RUN apt-get install -y net-tools iproute2

CMD s2dnglobby
