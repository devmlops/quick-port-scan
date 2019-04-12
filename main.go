package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type OpenAddresses struct {
	amu       sync.Mutex
	addresses map[string]string
}

func (a *OpenAddresses) addAddress(address, status string) {
	a.amu.Lock()
	a.addresses[address] = status
	a.amu.Unlock()
}

type QuickPortScan struct {
	ips     []string
	ports   []string
	timeout time.Duration

	OpenAddresses
	qmu                sync.Mutex
	isTooManyOpenFiles bool
}

func (q *QuickPortScan) addIPs(ips ...string) {
	for _, ip := range ips {
		q.ips = append(q.ips, ip)
	}
}

func (q *QuickPortScan) addPort(port int) {
	p := strconv.Itoa(port)
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

//func (q *QuickPortScan) ScanPorts(address []string) error {
//	for port := 1; port <= 8000; port++ {
//		address := net.JoinHostPort(ip, strconv.Itoa(port))
//		go q.isOpenedPort(address, timeout, &waitGroup, addAddress)
//	}
//}
//
//func (q *QuickPortScan) isOpenedPort(host, port string) {
//	defer w.Done()
//	conn, err := net.DialTimeout("tcp", address, timeout)
//	fmt.Println(err)
//	if oerr, ok := err.(*net.OpError); ok {
//		q.qmu.Lock()
//		fmt.Printf("-%#v\n", err)
//		fmt.Println(err)
//		if soerr, ok := oerr.Err.(*os.SyscallError); ok {
//			if soerr.Err == syscall.ECONNREFUSED {
//				fmt.Println("connection refused")
//			} else if soerr.Err == syscall.EMFILE {
//				panic("Too many open files. ")
//			}
//		} else if oerr.Timeout() {
//			fmt.Printf("i/o timeout")
//		}
//		q.qmu.Unlock()
//		return
//	} else if err != nil {
//		q.qmu.Lock()
//		fmt.Printf(" %#v %s\n", err, err)
//		q.qmu.Unlock()
//		return
//	}
//	defer conn.Close()
//	q.addAddress(address)
//}

func main() {
	var q QuickPortScan
	ips := []string{"127.0.0.1"}
	q.addIPs(ips...)
	q.addPortsRange(1, 10000)
	q.timeout = 10 * time.Second

	//var waitGroup sync.WaitGroup
	//waitGroup.Add(portsAmount)
	//
	//var addresses []string
	//addAddress := AddAddress(addresses)
	//
	//waitGroup.Wait()

	fmt.Printf("%#v", q)
}
