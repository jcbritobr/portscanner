package scanner

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

const (
	StClose = "closed"
	StOpen  = "open"
	//PtTCP is a ptrotocol used by scannner object and may be exported
	PtTCP                    = "tcp"
	ptFTP                    = "ftp"
	ptSSH                    = "ssh"
	ptTelnet                 = "telnet"
	ptSMTP                   = "smtp"
	ptOracleSQLNet           = "Oracle SQL*NET?"
	ptTFTP                   = "tftp"
	ptHTTP                   = "http"
	ptKerberos               = "kerberos"
	ptPop2                   = "pop2"
	ptPop3                   = "pop3"
	ptNTP                    = "ntp"
	ptNetbios                = "netbios"
	ptHTTPS                  = "https"
	ptSamba                  = "samba"
	ptCups                   = "cups"
	ptVNCRemoteDesktop       = "vnc remote desktop"
	prIRC                    = "irc"
	ptSQLService             = "sql service?"
	ptSQLNet                 = "sql net?"
	ptMicrosoftSQLServer     = "microsoft sql server"
	ptMicrosoftSQLMonitor    = "microsoft sql monitor"
	ptMySQL                  = "mysql"
	ptNovellNDPSPrinterAgent = "novell ndps printer agent"
	ptSMTPAlternate          = "smtp (alternate)"
	ptRTSP                   = "rtsp"
	ptCassandra              = "cassandra"
	ptMongoDB                = "mongodb <http://www.mongodb.org/>"
	ptMongoDBWebAdmin        = "mongodb web admin <http://www.mongodb.org/>"
	typeUnknown              = "<unknown>"

	errNetworkAddress = errMessage("Can't create a network address")
)

var knownPorts = map[int]string{
	21:    ptFTP,
	22:    ptSSH,
	23:    ptTelnet,
	25:    ptSMTP,
	66:    ptOracleSQLNet,
	69:    ptTFTP,
	80:    ptHTTP,
	88:    ptKerberos,
	109:   ptPop2,
	110:   ptPop3,
	123:   ptNTP,
	137:   ptNetbios,
	139:   ptNetbios,
	443:   ptHTTPS,
	445:   ptSamba,
	631:   ptCups,
	5800:  ptVNCRemoteDesktop,
	194:   prIRC,
	118:   ptSQLService,
	150:   ptSQLNet,
	1433:  ptMicrosoftSQLServer,
	1434:  ptMicrosoftSQLMonitor,
	3306:  ptMySQL,
	3396:  ptNovellNDPSPrinterAgent,
	3535:  ptSMTPAlternate,
	554:   ptRTSP,
	9160:  ptCassandra,
	27017: ptMongoDB,
	28017: ptMongoDBWebAdmin,
}

type errMessage string

func (e errMessage) Error() string {
	return string(e)
}

// Data holds the port information retrieved from the server
type Data struct {
	Port     int    `json:"port"`
	Status   string `json:"status"`
	Protocol string `json:"protocol"`
}

// Scanner implements a concurrent port scanner
type Scanner struct {
	ip      string
	workers int
	start   int
	end     int
	timeout uint16
}

// SortDataSlice is used to sort slice results
func SortDataSlice(slice []Data) {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].Port < slice[j].Port
	})
}

// NewScanner creates a new scanner object
// start is the port begining scan value that scanner will starts scan all ports,
// and end is the limit for port scanner. The workers param means the number
// of goroutines, and ip, the address for scan
func NewScanner(start, end, workers int, ip string, timeout uint16) *Scanner {
	return &Scanner{start: start, end: end, workers: workers, ip: ip, timeout: timeout}
}

// openConn will open a tcp connection with host and may return an error
// or a tcp connection
func (s *Scanner) openConn(host string) (net.Conn, error) {
	addr, err := net.ResolveTCPAddr(PtTCP, host)
	if err != nil {
		return nil, err
	}

	duration := time.Duration(s.timeout) * time.Millisecond
	conn, err := net.DialTimeout(PtTCP, addr.String(), duration)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Process starts the scan job
func (s *Scanner) Process() <-chan Data {
	buffer := []<-chan Data{}
	g := s.generate()

	for i := 0; i < s.workers; i++ {
		buffer = append(buffer, s.scanPort(g))
	}

	c := s.merge(buffer...)
	return c
}

// predictPort gets an int as parameter and return its respective
// port representation
func (s *Scanner) predictPort(port int) string {
	if rv, ok := knownPorts[port]; ok {
		return rv
	}
	return typeUnknown
}

// generate will build Data type objects and return a Data type buffered
// channel with Port field filled from the range of start field to end field
func (s *Scanner) generate() <-chan Data {
	c := make(chan Data, s.workers)
	go func() {
		for i := s.start; i <= s.end; i++ {
			d := Data{Port: i}
			c <- d
		}
		close(c)
	}()
	return c
}

// scanPort will try to open a tcp port and return a buffered channel with port status filled.
//
// buffer is a Data type buffered channel with ip and port filled.
//
// This function returns a Data type buffered channel with Status field filled.
func (s *Scanner) scanPort(buffer <-chan Data) <-chan Data {
	c := make(chan Data, s.workers)
	go func() {
		for dataItem := range buffer {
			host := fmt.Sprintf("%s:%d", s.ip, dataItem.Port)
			conn, err := s.openConn(host)
			if err != nil {
				dataItem.Status = StClose
			} else {
				dataItem.Status = StOpen
				dataItem.Protocol = s.predictPort(dataItem.Port)
				conn.Close()
			}
			c <- dataItem
		}
		close(c)
	}()

	return c
}

// merge will merge all channels in one.
//
// buffer is a Data slice or a sequence of Data separated by commas.
//
// It returns a Data channel with all input channels merged.
func (s *Scanner) merge(buffer ...<-chan Data) <-chan Data {
	var wg sync.WaitGroup
	out := make(chan Data, s.workers)
	output := func(c <-chan Data) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}

	wg.Add(len(buffer))
	for _, c := range buffer {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
