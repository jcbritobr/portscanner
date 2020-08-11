package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jcbritobr/portscanner/scanner"
)

func main() {
	start := flag.Int("start", 1, "-start <int value> is the start of port range")
	end := flag.Int("end", 80, "-end <int value> is the end of port range")
	timeout := flag.Int("t", 100, "-t <int value> is the value of connection timeout used")
	workers := flag.Int("workers", 1, "-workers <string value> is the number of concurrent process")
	host := flag.String("host", "127.0.0.1", "-host <string value> is the target host to be scaned")
	protocol := flag.String("proto", "tcp", "-proto <string value> is the protocol used - tcp or udp")

	flag.Parse()

	if scanner.PtTCP != *protocol && scanner.PtUDP != *protocol {
		fmt.Fprintf(os.Stderr, "Wrong protocol used %s. Use tcp or udp", *protocol)
	}

	s := scanner.NewScanner(*start, *end, *workers, *host, *protocol, uint16(*timeout))
	c := s.Process()

	for data := range c {
		fmt.Fprintf(os.Stdout, "port: %v	service: %s	status: %s\n", data.Port, data.Service, data.Status)
	}

}
