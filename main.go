package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jcbritobr/portscanner/scanner"
)

const (
	strMarker = "."
)

func main() {
	start := flag.Int("start", 1, "-start is the start of port range")
	end := flag.Int("end", 80, "-end is the end of port range")
	timeout := flag.Int("t", 100, "-t is the value of connection timeout used")
	workers := flag.Int("workers", 1, "-workers is the number of concurrent process")
	host := flag.String("host", "127.0.0.1", "-host is the target host to be scaned")
	protocol := flag.String("proto", "tcp", "-proto is the protocol used - tcp or udp")

	flag.Parse()

	if scanner.PtTCP != *protocol && scanner.PtUDP != *protocol {
		fmt.Fprintf(os.Stderr, "Wrong protocol used %s. Use tcp or udp", *protocol)
	}

	s := scanner.NewScanner(*start, *end, *workers, *host, *protocol, uint16(*timeout))
	c := s.Process()

	buffer := []scanner.Data{}
	fmt.Println("Generating report")
	for data := range c {
		fmt.Print(strMarker)
		buffer = append(buffer, data)
	}

	fmt.Printf("\n\n")

	scanner.SortDataSlice(buffer)

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 10, 3, ' ', 0)
	fmt.Fprintln(w, "Port\tProtocol\tStatus\t")
	fmt.Fprintln(w, "------\t--------\t------\t")
	for _, data := range buffer {
		fmt.Fprintln(w, fmt.Sprintf("%d\t%s\t%s\t", data.Port, data.Protocol, data.Status))
	}
	w.Flush()
}
