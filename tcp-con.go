package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s ip-addr\n", os.Args[0])
		os.Exit(1)
	}

	name := os.Args[1]
	tcpAddr, err := net.ResolveTCPAddr("tcp4", name)
	if err != nil {
		fmt.Printf("err is %v", err)
	}
	//addr := net.ParseIP(name)
	fmt.Println("The address is ", tcpAddr.String())
	fmt.Printf("The address is %v", tcpAddr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("err is %v", err)
	}
	data := []byte("HEAD / HTTP/1.0\r\n\r\n")

	_, err = conn.Write(data)
	if err != nil {
		fmt.Printf("err is %v", err)
	}

	readData, err := ioutil.ReadAll(conn)
	if err != nil {
		fmt.Printf("err is %v", err)
	}
	fmt.Printf("data is %s", string(readData))
	os.Exit(0)
}
