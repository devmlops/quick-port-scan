package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func NewQuickPortScan() *QuickPortScan {
	return &QuickPortScan{FoundAddresses: FoundAddresses{addresses: make(map[string]string)}}
}

type FoundAddresses struct {
	amu       sync.Mutex
	addresses map[string]string
}

func (a *FoundAddresses) addAddress(address, status string) {
	a.amu.Lock()
	a.addresses[address] = status
	a.amu.Unlock()
}

type QuickPortScan struct {
	ips     []string
	ports   []int
	timeout time.Duration
	threads int

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
	for _, ip := range q.ips {
		for _, port := range q.ports {
			fmt.Println(port)
			threads <- true
			address := net.JoinHostPort(ip, strconv.Itoa(port))
			go func(address string) {
				q.scanAddress(address)
				fmt.Println(address)
				<-threads
			}(address)
			fmt.Println(address)
		}
	}
	if q.isTooManyOpenFiles {
		return fmt.Errorf("Too many open files")
	}
	return nil
}

func (q *QuickPortScan) scanAddress(address string) {
	var status string
	conn, err := net.DialTimeout("tcp", address, q.timeout)
	if oerr, ok := err.(*net.OpError); ok {
		if soerr, ok := oerr.Err.(*os.SyscallError); ok {
			if soerr.Err == syscall.ECONNREFUSED {
				fmt.Println("connection refused")
				return
			} else if soerr.Err == syscall.EMFILE {
				q.setFlagTooManyOpenFiles(true)
				return
			}
		} else if oerr.Timeout() {
			status = "i/o timeout"
		}
	} else if err != nil {
		fmt.Printf(" %#v %s\n", err, err)
		return
	}
	defer conn.Close()
	status = "open"
	q.FoundAddresses.addAddress(address, status)
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
	fmt.Printf("%#v", q.FoundAddresses)
}
