## A Golang Port Scanner from scratch, using TDD
This is a simple **port scanner** that is able to map a large range of protocols. Using good design patterns for concurrency, like **fan-in/fan-out**, fixtures, goldenfiles and test helpers for tests, we create nice unit tests and acquire good performance in scanning process. All pakcages are covered by **TDD**, even when using tcp protocol for tests.

* **Install**
```
$ go install github.com/jcbritobr/portscanner@latest
```
* **Build**
```sh
$ go build -v
```

* **Test** \
The tests are using patterns like **golden files**, **fixtures** and **test helpers**. The golden files have inside it all output scenarios.
If tests failure happens in diffrent machines(because of port setup), just update the golden files.

```sh
$ go test ./... -cover -update
```
then test portscanner again

```sh
$ go test ./... -cover
```

* **Usage**
```sh
$ ./portscanner -h
Usage of ./portscanner:
  -end int
        -end is the end of port range (default 80)
  -host string
        -host is the target host to be scaned (default "127.0.0.1")
  -start int
        -start is the start of port range (default 1)
  -t int
        -t is the value of connection timeout used (default 100)
  -workers int
        -workers is the number of concurrent process (default 1)
```

```sh
$ ./portscanner -start 1 -end 2000 -workers 5 -host www.google.com
Generating report
processed: 100%

Port      Protocol  Status    

80        http      open      
443       https     open      
```