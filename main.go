package main

import (
	"flag"
	"fmt"
	"github.com/jcbritobr/portscanner/scanner"
)

func main() {
	start := flag.Int("start", 1, "-start is the start of port range")
	end := flag.Int("end", 80, "-end is the end of port range")
	timeout := flag.Int("t", 100, "-t is the value of connection timeout used")
	workers := flag.Int("workers", 1, "-workers is the number of concurrent process")
	host := flag.String("host", "127.0.0.1", "-host is the target host to be scaned")

	flag.Parse()

	counter := 0
	total := *end - *start

	s := scanner.NewScanner(*start, *end, *workers, *host, uint16(*timeout))
	c := s.Process()

	buffer := []scanner.Data{}
	fmt.Println("Generating report")

	for data := range c {
		buffer = append(buffer, data)
		value := (float32(counter) / float32(total)) * 100

		fmt.Printf("processed: %d%%\r", int(value))
		counter++
	}

	fmt.Printf("\n\n")

	scanner.SortDataSlice(buffer)

	fmt.Printf("%-10s%-10s%-10s\n\n", "Port", "Protocol", "Status")
	for _, data := range buffer {
		fmt.Printf(fmt.Sprintf("%-10d%-10s%-10s\n", data.Port, data.Protocol, data.Status))
	}
}
