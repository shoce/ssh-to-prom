
# https://hub.docker.com/_/golang/tags
FROM golang:1.23.0 AS build
RUN mkdir -p /root/ssh-to-prom/
COPY *.go go.mod go.sum /root/ssh-to-prom/
WORKDIR /root/ssh-to-prom/
RUN go version
RUN go get -a -u -v
RUN ls -l -a
RUN go build -o ssh-to-prom .
RUN ls -l -a


# https://hub.docker.com/_/alpine/tags
FROM alpine:3.20.2
RUN apk add --no-cache tzdata
RUN apk add --no-cache gcompat && ln -s -f -v ld-linux-x86-64.so.2 /lib/libresolv.so.2
RUN mkdir -p /opt/ssh-to-prom/
COPY --from=build /root/ssh-to-prom/ssh-to-prom /opt/ssh-to-prom/ssh-to-prom
RUN ls -l -a /opt/ssh-to-prom/
WORKDIR /opt/ssh-to-prom/
ENTRYPOINT ["./ssh-to-prom"]

