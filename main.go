package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type CheckState struct {
	requestsInProgress int
	toHandle           uint32 // count of required to handle
	handled            uint32 // count of already handle
	mutex              sync.RWMutex
	successResults     []PortProtocolCheckResult
	errorResults       []PortProtocolCheckResult
}

type PortProtocolCheckResult struct {
	resultType string // success, error
	port       int
	protocol   string
	message    string
}

type PortCheckResult struct {
	successCheck *PortProtocolCheckResult
	failedChecks []PortProtocolCheckResult
}

type Probes struct {
	http      bool
	https     bool
	memcached bool
}

var userAgents = []string{
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.0) Opera 12.14",
	"Mozilla/5.0 (X11; Ubuntu; Linux i686; rv:26.0) Gecko/20100101 Firefox/26.0",
	"Mozilla/5.0 (X11; U; Linux x86_64; en-US; rv:1.9.1.3) Gecko/20090913 Firefox/3.5.3",
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; en; rv:1.9.1.3) Gecko/20090824 Firefox/3.5.3 (.NET CLR 3.5.30729)",
	"Mozilla/5.0 (Windows NT 6.2) AppleWebKit/535.7 (KHTML, like Gecko) Comodo_Dragon/16.1.1.0 Chrome/16.0.912.63 Safari/535.7",
	"Mozilla/5.0 (Windows; U; Windows NT 5.2; en-US; rv:1.9.1.3) Gecko/20090824 Firefox/3.5.3 (.NET CLR 3.5.30729)",
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.1) Gecko/20090718 Firefox/3.5.1",
	"Mozilla / 5.0(X11;Linux i686; rv:81.0) Gecko / 20100101 Firefox / 81.0",
	"Mozilla / 5.0(Linuxx86_64;rv:81.0) Gecko / 20100101Firefox / 81.0",
	"Mozilla / 5.0(X11;Ubuntu;Linuxi686;rv:81.0) Gecko / 20100101Firefox / 81.0",
	"Mozilla / 5.0(X11;Ubuntu;Linuxx86_64;rv:81.0) Gecko / 20100101Firefox / 81.0",
	"Mozilla / 5.0(X11;Fedora;Linuxx86_64;rv:81.0) Gecko / 20100101Firefox / 81.0",
}

func main() {
	log.SetOutput(os.Stderr)

	// cli arguments
	var ipRangePattern = flag.String("ipRange", "", "IP Address Range")
	var probesString = flag.String("probes", "", "List of probes delimited by space")
	var maxConcurrentRequestCount = flag.Int("concurrent", 100, "Concurent request count")
	var verbose = flag.Bool("verbose", false, "Verbose")
	var portRangeString = flag.String("portRange", "1-65535", "Port range")
	flag.Parse()

	// ip range
	ipRange, err := NewIpRange(*ipRangePattern)
	if err != nil {
		log.Fatalln(err)
	}

	// port ragne
	portRange := portPatternToRange(*portRangeString)

	// probes
	probes := probesStringToProbes(*probesString)

	// init state
	var checkState = CheckState{
		requestsInProgress: 0,
		toHandle:           uint32(portRange.Size()+1) * (ipRange.Size() + 1),
		handled:            0,
		successResults:     []PortProtocolCheckResult{},
		errorResults:       []PortProtocolCheckResult{},
	}

	var waitGroup sync.WaitGroup

	resultChannel := make(chan PortCheckResult)
	go listenPortCheckResult(&checkState, resultChannel, &waitGroup)

	// start loop
	var currentIP net.IP
	for ipRange.Valid() {
		currentIP = ipRange.Current()

		for port := portRange.minPort; port <= portRange.maxPort; port++ {
			for {
				if checkState.requestsInProgress > *maxConcurrentRequestCount {
					time.Sleep(50 * time.Millisecond)
				} else {
					break
				}
			}

			checkState.mutex.Lock()
			checkState.requestsInProgress = checkState.requestsInProgress + 1
			checkState.mutex.Unlock()

			// find services on port
			waitGroup.Add(1)

			go probe(currentIP.String(), port, probes, resultChannel)
		}

		ipRange.Next()
	}

	waitGroup.Wait()

	fmt.Println("\n")

	// print errors
	if *verbose {
		if len(checkState.errorResults) > 0 {
			printResult("Errors", checkState.errorResults)
		}
	}

	// print results
	printResult("Found services", checkState.successResults)
}

func probe(ip string, port int, probes Probes, resultChannel chan PortCheckResult) {
	var protocolCheckResult PortProtocolCheckResult
	var failedChecks []PortProtocolCheckResult

	// check http
	if probes.http {
		protocolCheckResult = checkIpHasHttpService("http", ip, port)
		if protocolCheckResult.resultType == "success" {
			resultChannel <- PortCheckResult{
				successCheck: &protocolCheckResult,
			}

			return
		} else {
			failedChecks = append(failedChecks, protocolCheckResult)
		}
	}

	// check https
	if probes.https {
		protocolCheckResult = checkIpHasHttpService("https", ip, port)
		if protocolCheckResult.resultType == "success" {
			resultChannel <- PortCheckResult{
				successCheck: &protocolCheckResult,
			}

			return
		} else {
			failedChecks = append(failedChecks, protocolCheckResult)
		}
	}

	// check memcached
	if probes.memcached {
		protocolCheckResult = checkIpHasMemcachedService(ip, port)
		if protocolCheckResult.resultType == "success" {
			resultChannel <- PortCheckResult{
				successCheck: &protocolCheckResult,
			}

			return
		} else {
			failedChecks = append(failedChecks, protocolCheckResult)
		}
	}

	resultChannel <- PortCheckResult{
		failedChecks: failedChecks,
	}
}

func printResult(title string, result []PortProtocolCheckResult) {
	if len(result) == 0 {
		return
	}

	fmt.Println("")
	fmt.Println(title + ":")
	for i := 0; i < len(result); i++ {
		fmt.Println(result[i].message)
	}
}

func listenPortCheckResult(
	checkState *CheckState,
	resultChannel chan PortCheckResult,
	waitGroup *sync.WaitGroup,
) {
	for {
		result := <-resultChannel
		waitGroup.Done()

		// add state
		checkState.mutex.Lock()
		checkState.requestsInProgress = checkState.requestsInProgress - 1
		checkState.handled = checkState.handled + 1
		if result.successCheck != nil {
			checkState.successResults = append(checkState.successResults, *result.successCheck)
		} else if len(result.failedChecks) > 0 {
			checkState.errorResults = append(checkState.errorResults, result.failedChecks...)
		}

		checkState.mutex.Unlock()

		// show progress
		fmt.Printf(
			"\r[%3d%%][%d/%d] Errors: %d, Found: %d",
			int64(float32(checkState.handled)/float32(checkState.toHandle)*100),
			checkState.handled,
			checkState.toHandle,
			len(checkState.errorResults),
			len(checkState.successResults),
		)
	}
}

func probesStringToProbes(probesString string) Probes {
	if probesString == "" {
		return Probes{
			http:      true,
			https:     true,
			memcached: true,
		}
	}

	probes := Probes{
		http:      false,
		https:     false,
		memcached: false,
	}

	probesSlice := strings.Split(probesString, " ")
	for _, probe := range probesSlice {
		switch probe {
		case "http":
			probes.http = true
			break
		case "https":
			probes.https = true
			break
		case "memcached":
			probes.memcached = true
			break
		}
	}

	return probes
}

func portPatternToRange(portRangeString string) PortRange {
	portRange := strings.Split(portRangeString, "-")
	minPort := 1
	maxPort := 65535
	if len(portRange) == 1 {
		minPort, _ = strconv.Atoi(portRange[0])
		maxPort, _ = strconv.Atoi(portRange[0])
	} else {
		if portRange[0] == "" {
			portRange[0] = "1"
		}

		if portRange[1] == "" {
			portRange[1] = "65535"
		}

		minPort, _ = strconv.Atoi(portRange[0])
		maxPort, _ = strconv.Atoi(portRange[1])
	}

	if minPort > maxPort {
		log.Fatalln("Invalid port range")
	}

	return PortRange{minPort: minPort, maxPort: maxPort}
}

func checkIpHasHttpService(
	schema string,
	ip string,
	port int,
) PortProtocolCheckResult {
	var userAgent = userAgents[rand.Intn(len(userAgents))]

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	httpUrl := schema + "://" + ip + ":" + strconv.Itoa(port) + "/"
	request, _ := http.NewRequest("GET", httpUrl, nil)
	request.Header.Set("User-Agent", userAgent)
	response, err := client.Do(request)

	if err == nil {
		return PortProtocolCheckResult{
			port:       port,
			protocol:   schema,
			resultType: "success",
			message:    schema + "\t" + httpUrl + "\t" + strconv.Itoa(response.StatusCode),
		}
	} else {
		return PortProtocolCheckResult{
			port:       port,
			protocol:   schema,
			resultType: "error",
			message:    err.Error(),
		}
	}
}

func checkIpHasMemcachedService(
	ip string,
	port int,
) PortProtocolCheckResult {
	addr := ip + ":" + strconv.Itoa(port)

	mc := memcache.New(addr)
	mc.Timeout = 1 * time.Second

	//err := mc.Set(&memcache.Item{Key: "lol", Value: []byte("kek")})
	err := mc.Ping()

	if err == nil {
		return PortProtocolCheckResult{
			port:       port,
			protocol:   "memcached",
			resultType: "success",
			message:    "memcached\t" + ip + ":" + strconv.Itoa(port),
		}
	} else {
		return PortProtocolCheckResult{
			port:       port,
			protocol:   "memcached",
			resultType: "error",
			message:    err.Error(),
		}
	}
}
