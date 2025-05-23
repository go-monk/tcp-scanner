package main

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"
)

var (
	hostToScan     = "scanme.nmap.org"
	portRangeStart = 1
	portRangeEnd   = 1024
	nWorkers       = 100
	connTimeout    = 1 * time.Second
)

func worker(portsToScan, portsScanned chan int) {
	for port := range portsToScan {
		addr := net.JoinHostPort(hostToScan, strconv.Itoa(port))
		conn, err := net.DialTimeout("tcp", addr, connTimeout)
		if err != nil {
			// closed port		syn->, <-rst
			// filtered port	syn->, timeout
			portsScanned <- 0
			continue
		}
		conn.Close()
		portsScanned <- port
	}
}

func main() {
	portsToScan := make(chan int, nWorkers) // can hold nWorkers items before sender blocks
	portsScanned := make(chan int)

	for i := 0; i < nWorkers; i++ {
		go worker(portsToScan, portsScanned)
	}

	go func() {
		for i := portRangeStart; i <= portRangeEnd; i++ {
			portsToScan <- i
		}
	}()

	var openPorts []int

	for i := portRangeStart; i <= portRangeEnd; i++ {
		port := <-portsScanned
		if port != 0 {
			openPorts = append(openPorts, port)
		}
		fmt.Printf("\r%d", i)
	}
	fmt.Printf("\n")

	sort.Ints(openPorts)
	fmt.Println(hostToScan, openPorts)
}
