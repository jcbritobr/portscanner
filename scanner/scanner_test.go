package scanner

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const (
	tstPort         = "3030"
	srvAddr         = ":3030"
	fsFixtureFolder = "testdata"
)

var (
	update = flag.Bool("update", false, "update golden files")
)

func checkErrorFatalf(t *testing.T, message string, err error) {
	if err != nil {
		t.Fatalf("%s err %v\n", message, err)
	}
}

func spawnServerHelper(t *testing.T, host, protocol string) net.Conn {
	t.Helper()
	listener, err := net.Listen(protocol, host)
	checkErrorFatalf(t, "cant create listener", err)

	var server net.Conn
	go func() {
		defer listener.Close()
		server, err = listener.Accept()
	}()

	return server
}

func openGoldenFileHelper(t *testing.T, filename, source string, update bool) string {
	t.Helper()
	path := filepath.Join(fsFixtureFolder, filename)
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	defer file.Close()
	checkErrorFatalf(t, fmt.Sprintf("cant open golden file %s", path), err)
	if update {
		err = os.Truncate(path, 0)
		checkErrorFatalf(t, fmt.Sprintf("cant truncate file %s", path), err)

		_, err = file.WriteString(source)
		checkErrorFatalf(t, fmt.Sprintf("cant write golden file %s", path), err)
		return source
	}

	content, err := ioutil.ReadAll(file)
	checkErrorFatalf(t, fmt.Sprintf("cant read golden file %s", path), err)
	return string(content)
}

func openDataFixtureHelper(t *testing.T, filename string) []Data {
	t.Helper()
	path := filepath.Join(fsFixtureFolder, filename)
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	defer file.Close()
	checkErrorFatalf(t, fmt.Sprintf("cant open fixture %s", path), err)

	content, err := ioutil.ReadAll(file)
	checkErrorFatalf(t, fmt.Sprintf("cant read file %s", path), err)

	var buffer []Data
	err = json.Unmarshal(content, &buffer)
	checkErrorFatalf(t, fmt.Sprintf("cant unmarshal json file %s", path), err)

	return buffer
}

func generateDataBuffer(start, end int, last Data) []Data {
	buffer := []Data{}
	for i := start; i < end; i++ {
		buffer = append(buffer, Data{Port: i, Status: stClose})
	}

	buffer = append(buffer, last)
	return buffer
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
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
		{
			"NewScanner A", args{start: 0, end: 100, workers: 1, ip: "192.168.1.1", protocol: PtTCP},
			&Scanner{start: 0, end: 100, workers: 1, ip: "192.168.1.1", protocol: PtTCP},
		},
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
		want string
	}{
		{"generate A", args{NewScanner(1, 3, 1, "192.168.1.1", PtTCP, 100)}, "generatea.golden"},
		{"generate B", args{NewScanner(1, 1, 1, "192.168.1.10", PtTCP, 100)}, "generateb.golden"},
		{"generate C", args{NewScanner(2, 3, 1, "192.168.1.10", PtTCP, 100)}, "generatec.golden"},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			got := []Data{}
			c := tC.args.scanner.generate()

			for d := range c {
				got = append(got, d)
			}

			str := fmt.Sprintf("%v", got)
			want := openGoldenFileHelper(t, tC.want, str, *update)

			if !reflect.DeepEqual(str, want) {
				t.Errorf("generate() = %v, want %v", got, want)
			}
		})
	}
}

func TestScanPort(t *testing.T) {
	_ = spawnServerHelper(t, ":3015", PtTCP)
	type args struct {
		scan *Scanner
	}
	testCases := []struct {
		name string
		args args
		want string
	}{
		{"scanPort A", args{NewScanner(3000, 3030, 2, "localhost", PtTCP, 100)}, "scanporta.golden"},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			result := []Data{}
			c := tC.args.scan.generate()
			for dataResult := range tC.args.scan.scanPort(c) {
				result = append(result, dataResult)
			}

			got := fmt.Sprintf("%v", result)
			want := openGoldenFileHelper(t, tC.want, got, *update)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("scanPort() = %s want %v", got, want)
			}
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

	type args struct {
		scanner *Scanner
	}
	testCases := []struct {
		name string
		args args
		want string
	}{
		{"processA", args{scanner: NewScanner(3000, 3030, 3, "localhost", PtTCP, 100)}, "processa.golden"},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			result := []Data{}
			c := tC.args.scanner.Process()
			for data := range c {
				result = append(result, data)
			}
			SortDataSlice(result)
			got := fmt.Sprintf("%v", result)
			want := openGoldenFileHelper(t, tC.want, got, *update)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("Process() = %v want %v", got, want)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	_ = spawnServerHelper(t, ":2001", PtTCP)
	type args struct {
		scanner *Scanner
	}
	testCases := []struct {
		name string
		args args
		want string
	}{
		{"mergeA", args{scanner: NewScanner(3030, 3030, 1, "localhost", PtTCP, 100)}, "mergea.golden"},
		{"mergeB", args{scanner: NewScanner(2000, 3030, 5, "localhost", PtTCP, 100)}, "mergeb.golden"},
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
			got := fmt.Sprintf("%v", result)
			want := openGoldenFileHelper(t, tC.want, got, *update)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("merge() = %v want %v", got, want)
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
		want string
	}{
		{"sortDataSliceA", args{openDataFixtureHelper(t, "fixsortdataslice.json")}, "sortdataslicea.golden"},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {

			SortDataSlice(tC.args.slice)
			got := fmt.Sprintf("%v", tC.args.slice)
			want := openGoldenFileHelper(t, tC.want, got, *update)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("sortDataSlice() = %v want %v", got, want)
			}
		})
	}
}
