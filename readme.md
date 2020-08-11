## A Golang Port Scanner from scratch, using TDD
This is a simple **port scanner** that is able to map a large range of protocols. Using good design patterns for concurrency, like **fan-in/fan-out**, we acquire good performance in scanning process. All pakcages are covered by **TDD**, even when using tcp protocol for tests.

* **UDP isn't implemented yet**

* **Build**
```shell
$ go build -v
```

* **Usage**
```shell
$ ./portscanner -h
Usage of ./portscanner:
  -end int
        -end <int value> is the end of port range (default 80)
  -host string
        -host <string value> is the target host to be scaned (default "127.0.0.1")
  -proto string
        -proto <string value> is the protocol used - tcp or udp (default "tcp")
  -start int
        -start <int value> is the start of port range (default 1)
  -t int
        -t <int value> is the value of connection timeout used (default 100)
  -workers int
        -workers <string value> is the number of concurrent process (default 1)
```

```shell
$ ./portscanner -end 100 -host www.google.com | grep open
port: 80        service: http   status: open
```