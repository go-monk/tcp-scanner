It's quite common to need to discover which open (listening) ports exist on a remote system. This requirement can arise for security professionals during the reconnaissance phase of a penetration test. DevOps engineers might also need to determine what network services a system exposes—without logging into it.

Let's build a tool to do just that. (Of course, several tools already exist for this, like nmap.) We'll focus specifically on the TCP protocol here.

First, we create two channels (the first one is buffered) and a pool of workers that will perform the scanning by attempting to connect to ports:

```go
func main() {
	portsToScan := make(chan int, nWorkers)
	portsScanned := make(chan int)

	for i := 0; i < nWorkers; i++ {
		go worker(portsToScan, portsScanned)
	}
}

func worker(portsToScan, portsScanned chan int) {
	for port := range portsToScan {
		addr := net.JoinHostPort(hostToScan, strconv.Itoa(port))
		conn, err := net.DialTimeout("tcp", addr, connTimeout)
		if err != nil {
			// Closed port:     SYN ->, <- RST
			// Filtered port:   SYN ->, timeout
			portsScanned <- 0
			continue
		}
		conn.Close()
		portsScanned <- port
	}
}
```

Each worker takes port numbers from the `portsToScan` channel and attempts to connect to them. If the connection fails, it sends `0` to the `portsScanned` channel. Otherwise, it closes the connection and sends the open port number to `portsScanned`.

Next, we enqueue a range of ports for the workers to scan. Note that ports numbered ≤ 1024 are typically reserved for services requiring root privileges. We launch this loop in a goroutine so it doesn’t block the main thread:

```go
go func() {
	for i := portRangeStart; i <= portRangeEnd; i++ {
		portsToScan <- i
	}
}()
```

Now we collect the results sent back by the workers. If the result is not zero, we know the port is open and add it to the `openPorts` slice. We also print the progress by displaying the number of ports scanned:

```go
var openPorts []int

for i := portRangeStart; i <= portRangeEnd; i++ {
	port := <-portsScanned
	if port != 0 {
		openPorts = append(openPorts, port)
	}
	fmt.Printf("\r%d", i)
}
```

Finally, we sort and print the open ports:

```go
sort.Ints(openPorts)
fmt.Println(host, openPorts)
```

To run the code:

```sh
❯ go run tcpscanner.go 
1024
scanme.nmap.org [22 80]
```

**NOTE:** Only scan hosts that you are explicitly authorized to scan. You are permitted to scan `scanme.nmap.org`—see [http://scanme.nmap.org](http://scanme.nmap.org) for details.

---

Adapted from "Black Hat Go" by Tom Steele, Chris Patten, and Dan Kottmann (2020).
