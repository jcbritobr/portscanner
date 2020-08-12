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
$ ./portscanner -workers 10 -t 1000 -end 80 -host www.mane.com | grep open
21       ftp        open
22       ssh        open
80       http       open
```

```shell
$ ./portscanner.exe -t 1000 -workers 5 -host www.mane.com -end 80
Generating report
................................................................................

Port     Service    Status
------   --------   ------
1                   closed
2                   closed
3                   closed
4                   closed
5                   closed
6                   closed
7                   closed
8                   closed
9                   closed
10                  closed
11                  closed
12                  closed
13                  closed
14                  closed   
15                  closed
16                  closed
17                  closed
18                  closed
19                  closed
20                  closed
21       ftp        open
22       ssh        open
23                  closed
24                  closed
25                  closed
26                  closed
27                  closed
28                  closed
29                  closed   
30                  closed
31                  closed
32                  closed
33                  closed
34                  closed
35                  closed
36                  closed
37                  closed
38                  closed
39                  closed
40                  closed
41                  closed
42                  closed
43                  closed
44                  closed
45                  closed
46                  closed
47                  closed
48                  closed
49                  closed
50                  closed
51                  closed   
52                  closed
53                  closed
54                  closed
55                  closed
56                  closed
57                  closed
58                  closed
59                  closed
60                  closed
61                  closed
62                  closed   
63                  closed
64                  closed
65                  closed
66                  closed
67                  closed
68                  closed
69                  closed
70                  closed
71                  closed
72                  closed
73                  closed
74                  closed
75                  closed
76                  closed
77                  closed
78                  closed
79                  closed
80       http       open
```