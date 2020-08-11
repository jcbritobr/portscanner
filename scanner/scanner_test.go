package scanner

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
)

const (
	tstPort = "3030"
	srvAddr = ":3030"
)

type server interface {
	run() error
	close() error
}

type tcpServer struct {
	addr   string
	server net.Listener
}

func newServer(protocol, addr string) (server, error) {
	switch strings.ToLower(protocol) {
	case PtTCP:
		return &tcpServer{addr: addr}, nil
	case PtUDP:
	}

	return nil, fmt.Errorf("Invalid protocol given")
}

func (t *tcpServer) close() error {
	return t.server.Close()
}

func (t *tcpServer) run() error {
	server, err := net.Listen("tcp", t.addr)
	t.server = server
	if err != nil {
		return err
	}

	for {
		conn, err := t.server.Accept()
		if err != nil {
			return fmt.Errorf("cant accept the connection %v", err)
		}
		go t.handleClient(conn)
	}
}

func (t *tcpServer) handleClient(conn net.Conn) {
	defer conn.Close()
}

func checkErrorPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func generateDataBuffer(start, end int, last Data) []Data {
	buffer := []Data{}
	for i := start; i < end; i++ {
		buffer = append(buffer, Data{Port: i, Status: stClose})
	}

	buffer = append(buffer, last)
	return buffer
}

func init() {
	srv, err := newServer(PtTCP, srvAddr)
	checkErrorPanic(err)
	go func() {
		srv.run()
	}()
}

//------------------------------------------------ Start test functions ------------------------------------------------//

func TestNewScanner(t *testing.T) {
	type args struct {
		start    int
		end      int
		workers  int
		ip       string
		protocol string
		timeout  uint16
	}
	tests := []struct {
		name string
		args args
		want *Scanner
	}{
		{"NewScanner A", args{start: 0, end: 100, workers: 1, ip: "192.168.1.1", protocol: PtTCP}, &Scanner{start: 0, end: 100, workers: 1, ip: "192.168.1.1", protocol: PtTCP}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewScanner(tt.args.start, tt.args.end, tt.args.workers, tt.args.ip, tt.args.protocol, tt.args.timeout); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewScanner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	type args struct {
		scanner *Scanner
	}
	testCases := []struct {
		name string
		args args
		want []Data
	}{
		{"generate A", args{NewScanner(1, 3, 1, "192.168.1.1", PtTCP, 100)}, []Data{{Port: 1}, {Port: 2}, {Port: 3}}},
		{"generate B", args{NewScanner(1, 1, 1, "192.168.1.10", PtTCP, 100)}, []Data{{Port: 1}}},
		{"generate B", args{NewScanner(2, 3, 1, "192.168.1.10", PtTCP, 100)}, []Data{{Port: 2}, {Port: 3}}},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			got := []Data{}
			c := tC.args.scanner.generate()

			for d := range c {
				got = append(got, d)
			}

			if !reflect.DeepEqual(got, tC.want) {
				t.Errorf("generate() = %v, want %v", got, tC.want)
			}
		})
	}
}

func TestScanPort(t *testing.T) {
	type args struct {
		scan *Scanner
	}
	testCases := []struct {
		name string
		args args
		want Data
	}{
		{"scanPort A", args{NewScanner(3029, 3030, 1, "localhost", PtTCP, 100)}, Data{Port: 3030, Service: typeUnknown, Status: stOpen}},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			result := []Data{}
			c := tC.args.scan.generate()
			for dataResult := range tC.args.scan.scanPort(c) {
				result = append(result, dataResult)
			}

			for _, dataResult := range result {
				if dataResult.Port == tC.want.Port && dataResult.Service == tC.want.Service && dataResult.Status == tC.want.Status {
					return
				}
			}

			t.Errorf("scanPort() = not found want %v", tC.want)
		})
	}
}

func TestOpenConn(t *testing.T) {
	type args struct {
		addr string
	}
	testCases := []struct {
		name string
		args args
		want bool
	}{
		{"openConnA", args{addr: ":3030"}, true},
		{"openConnB", args{addr: ":4030"}, false},
		{"openConnC", args{addr: ":AAAA"}, false},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			scan := NewScanner(0, 0, 0, "", PtTCP, 100)
			_, err := scan.openConn(tC.args.addr)
			if err == nil && !tC.want {
				t.Errorf("openConn() = %v want %v", err, tC.want)
			}
		})
	}
}

func TestPredictPort(t *testing.T) {
	type args struct {
		port int
	}
	testCases := []struct {
		name string
		args args
		want string
	}{
		{"predictPortA", args{port: 21}, ptFTP},
		{"predictPortA", args{port: 4040}, typeUnknown},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			s := Scanner{}
			result := s.predictPort(tC.args.port)
			if result != tC.want {
				t.Errorf("predictPort() = %v want %v", result, tC.want)
			}
		})
	}
}

func TestError(t *testing.T) {
	testCases := []struct {
		name string
		want error
	}{
		{"errorA", errNetworkAddress},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			f := func() error {
				return fmt.Errorf("Bad network address %w", errNetworkAddress)
			}

			err := f()
			if !errors.Is(err, errNetworkAddress) {
				t.Errorf("error() = %v want %v", err, tC.want)
			}
		})
	}
}

func TestProcess(t *testing.T) {
	dataA := generateDataBuffer(3000, 3030, Data{Port: 3030, Service: typeUnknown, Status: stOpen})
	type args struct {
		scanner *Scanner
	}
	testCases := []struct {
		name string
		args args
		want []Data
	}{
		{"processA", args{scanner: NewScanner(3000, 3030, 3, "localhost", PtTCP, 100)}, dataA},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			result := []Data{}
			c := tC.args.scanner.Process()
			for data := range c {
				result = append(result, data)
			}
			SortDataSlice(result)

			if !reflect.DeepEqual(result, tC.want) {
				t.Errorf("Process() = %v want %v", result, tC.want)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	type args struct {
		scanner *Scanner
	}
	testCases := []struct {
		name string
		args args
		want []Data
	}{
		{"mergeA", args{scanner: NewScanner(3030, 3030, 1, "localhost", PtTCP, 100)}, []Data{{Port: 3030, Service: typeUnknown, Status: stOpen}}},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			buffer := []<-chan Data{}
			result := []Data{}
			c := tC.args.scanner.generate()

			for i := 0; i < tC.args.scanner.workers; i++ {
				buffer = append(buffer, tC.args.scanner.scanPort(c))
			}

			z := tC.args.scanner.merge(buffer...)

			for data := range z {
				result = append(result, data)
			}

			SortDataSlice(result)

			if !reflect.DeepEqual(result, tC.want) {
				t.Errorf("merge() = %v want %v", result, tC.want)
			}
		})
	}
}

func TestSortDataSlice(t *testing.T) {
	type args struct {
		slice []Data
	}
	testCases := []struct {
		name string
		args args
		want []Data
	}{
		{"sortDataSliceA", args{[]Data{{Port: 10, Service: "", Status: stClose}, {Port: 6, Service: "", Status: stClose}}}, []Data{{Port: 6, Service: "", Status: stClose}, {Port: 10, Service: "", Status: stClose}}},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {

			SortDataSlice(tC.args.slice)

			if !reflect.DeepEqual(tC.args.slice, tC.want) {
				t.Errorf("sortDataSlice() = %v want %v", tC.args.slice, tC.want)
			}
		})
	}
}
