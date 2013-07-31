package main

import (
	"net"
	"crypto/tls"
)

// defaultPort joins addr and port if addr contains just the host name or IP.
func defaultPort(addr, port string) string {
        _, _, err := net.SplitHostPort(addr)
        if err != nil {
                addr = net.JoinHostPort(addr, port)
        }
        return addr
}

// setServerName returns a new TLS configuration with ServerName set to host if
// the original configuration was nil or config.ServerName was empty.
func setServerName(config *tls.Config, host string) *tls.Config {
	if config == nil {
		config = &tls.Config{ServerName: host}
	} else if config.ServerName == "" {
		c := *config
		c.ServerName = host
		config = &c
	}
	return config
}
