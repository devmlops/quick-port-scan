package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var mu sync.Mutex

func AddAddress(addresses []string) func(string) {
	return func(address string) {
		addresses = append(addresses, address)
	}
}

func Dial(address string, timeout time.Duration, w *sync.WaitGroup, addAddress func(string)) {
	defer w.Done()
	conn, err := net.DialTimeout("tcp", address, timeout)
	fmt.Println(err)
	if oerr, ok := err.(*net.OpError); ok {
		mu.Lock()
		fmt.Printf("%#v\n", err)
		fmt.Println(err)
		if soerr, ok := oerr.Err.(*os.SyscallError); ok {
			if soerr.Err == syscall.ECONNREFUSED {
				fmt.Println("connection refused")
			} else if soerr.Err == syscall.EMFILE {
				fmt.Println("too many open files")
			}
		} else if oerr.Timeout() {
			fmt.Printf("i/o timeout")
		}
		mu.Unlock()
		return
	} else if err != nil {
		mu.Lock()
		fmt.Printf(" %#v %s\n", err, err)
		mu.Unlock()
		return
	}
	defer conn.Close()
	addAddress(address)
}

func main() {
	ip := "127.0.0.1"
	timeout := 10 * time.Second

	portsAmount := 8000
	var waitGroup sync.WaitGroup
	waitGroup.Add(portsAmount)

	var addresses []string
	addAddress := AddAddress(addresses)

	for port := 1; port <= 8000; port++ {
		address := net.JoinHostPort(ip, strconv.Itoa(port))
		go Dial(address, timeout, &waitGroup, addAddress)
	}
	waitGroup.Wait()
	fmt.Println(addresses)
}
