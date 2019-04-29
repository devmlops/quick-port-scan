package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func NewQuickPortScan() *QuickPortScan {
	return &QuickPortScan{
		FoundAddresses: FoundAddresses{addresses: make(map[string]map[int]string)},
		pauseTime:      0,
	}
}

type FoundAddresses struct {
	amu       sync.Mutex
	addresses map[string]map[int]string
}

func (a *FoundAddresses) addAddress(ip string, port int, status string) {
	a.amu.Lock()
	if _, ok := a.addresses[ip]; ok != true {
		a.addresses[ip] = map[int]string{}
	}
	a.addresses[ip][port] = status
	a.amu.Unlock()
}

func (a *FoundAddresses) PrintAddresses() {
	var rawIPs []string
	for address := range a.addresses {
		rawIPs = append(rawIPs, address)
	}
	sort.Strings(rawIPs)
	var sortedAddresses []string
	for _, ip := range rawIPs {
		for port := range a.addresses[ip] {
			address := fmt.Sprintf("%s:%d", ip, port)
			sortedAddresses = append(sortedAddresses, address)
		}
	}
	for _, addr := range sortedAddresses {
		fmt.Printf("%s\n", addr)
	}

}

type QuickPortScan struct {
	ips       []string
	ports     []int
	timeout   time.Duration
	pauseTime time.Duration
	threads   int

	FoundAddresses
	qmu                sync.Mutex
	isTooManyOpenFiles bool
}

func (q *QuickPortScan) addIPs(ips ...string) {
	for _, ip := range ips {
		q.ips = append(q.ips, ip)
	}
}

func (q *QuickPortScan) addPort(port int) {
	p := port
	q.ports = append(q.ports, p)
}

func (q *QuickPortScan) addPortsRange(fromPort, toPort int) {
	for port := fromPort; port <= toPort; port++ {
		q.addPort(port)
	}
}

func (q *QuickPortScan) setFlagTooManyOpenFiles(s bool) {
	q.qmu.Lock()
	q.isTooManyOpenFiles = s
	q.qmu.Unlock()
}

func (q *QuickPortScan) StartScan() error {
	q.setFlagTooManyOpenFiles(false)
	threads := make(chan bool, q.threads)
	var wg sync.WaitGroup
	wg.Add(len(q.ips) * len(q.ports))
	for _, ip := range q.ips {
		for _, port := range q.ports {
			threads <- true
			go func(ip string, port int) {
				q.scanAddress(ip, port)
				<-threads
				wg.Done()
			}(ip, port)
			time.Sleep(q.pauseTime)
		}
	}
	wg.Wait()
	if q.isTooManyOpenFiles {
		return fmt.Errorf("Too many open files")
	}
	return nil
}

func (q *QuickPortScan) scanAddress(ip string, port int) {
	var status string
	address := net.JoinHostPort(ip, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", address, q.timeout)
	if oerr, ok := err.(*net.OpError); ok {
		if soerr, ok := oerr.Err.(*os.SyscallError); ok {
			if soerr.Err == syscall.ECONNREFUSED {
				return
			} else if soerr.Err == syscall.EMFILE {
				q.setFlagTooManyOpenFiles(true)
				return
			}
		} else if oerr.Timeout() {
			status = "i/o timeout"
			return
		}
	} else if err != nil {
		fmt.Printf(" %#v %s\n", err, err)
		return
	}
	defer conn.Close()
	status = "open"
	q.FoundAddresses.addAddress(ip, port, status)
}

func main() {
	q := NewQuickPortScan()
	ips := []string{"127.0.0.1"}
	q.addIPs(ips...)
	q.addPortsRange(1, 65535)
	q.timeout = 10 * time.Second
	q.threads = 1000
	if err := q.StartScan(); err != nil {
		log.Fatalf("Too many open files")
	}
	q.PrintAddresses()
}
