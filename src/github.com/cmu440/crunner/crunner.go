package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	defaultHost = "localhost"
	defaultPort = 9999
)

// To test your server implementation, you might find it helpful to implement a
// simple 'client runner' program. The program could be very simple, as long as
// it is able to connect with and send messages to your server and is able to
// read and print out the server's response to standard output. Whether or
// not you add any code to this file will not affect your grade.
func main() {
	fmt.Println("Not implemented.")
	conn, err := net.Dial("tcp", "localhost:9999")
	if err != nil {
		fmt.Println("Failed to start: ", err.Error())
	}
	inputReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Please input a command, type Q to quit!")
		if err != nil {
			fmt.Println("Command illegal!")
		}
		input, _ := inputReader.ReadString('\n')
		command := strings.Trim(input, "\r\n")
		if command == "Q" {
			return
		}

        _, err := conn.Write([]byte(command))
        if err != nil {
            return
        }
	}
}
