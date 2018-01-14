package main

import "net"
import "fmt"
import "bufio"
import "strings" // only needed below for sample processing

type user struct {
	nick, pass string
	chans []string
}

func main() {

	fmt.Println("Launching server...")

	// listen on all interfaces
	ln, _ := net.Listen("tcp", ":8081")

	// accept connection on port
	con, _ := ln.Accept()
	sc := bufio.NewScanner(con)
	fmt.Println("New connection)")
	users := make(map[string]user)
	var n, p, u string
	for sc.Scan() {
		st := strings.Split(sc.Text(), " ")
		switch strings.ToLower(st[0]) {
			case "name":
				n = strings.Join(st[1:], " ")
				fmt.Println("It's nick : ", n)
			case "user":
				u = strings.Join(st[1:], " ")
				fmt.Println("It's user: ", u)
			default : fmt.Println("Nor user or name :(  : ", st)
		}
		fmt.Println(len(n), len(u))
		if len(n) > 0 && len(u) > 0 {
			users[u] = user{nick: n, pass: p}
			fmt.Println("potato")
			break
		}
	}
	for u, _ := range users {
		fmt.Println(":server 001 ", u, ": Hehey you're welcome")
	}
	// output message received
	//fmt.Print("Message Received:", string(message))
	// sample process for string received
	//newmessage := strings.ToUpper(message)
	// send new string back to client
	//conn[0].Write([]byte(newmessage + "\n"))
}
