
# https://hub.docker.com/_/golang/tags
FROM golang:1.25-alpine AS build
RUN mkdir -p /ssh-to-prom/
COPY *.go go.mod go.sum /ssh-to-prom/
WORKDIR /ssh-to-prom/
RUN go version
RUN go get -a -u -v
RUN ls -l -a
RUN go build -o ssh-to-prom .
RUN ls -l -a


# https://hub.docker.com/_/alpine/tags
FROM alpine:3
RUN apk add --no-cache tzdata gcompat && ln -s -f -v ld-linux-x86-64.so.2 /lib/libresolv.so.2
RUN mkdir -p /ssh-to-prom/
COPY --from=build /ssh-to-prom/ssh-to-prom /ssh-to-prom/ssh-to-prom
RUN ls -l -a /ssh-to-prom/
WORKDIR /ssh-to-prom/
ENTRYPOINT ["/ssh-to-prom/ssh-to-prom"]

