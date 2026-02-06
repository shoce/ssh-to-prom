package main

import (
	"net"
	"time"
)

type ConnEvent struct {
	Timestamp  time.Time
	Accepted   bool
	AuthMethod string
	User       string
	Addr       net.IP
	Port       int
}
