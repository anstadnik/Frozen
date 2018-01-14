package main

import (
	"net"
	"fmt"
	"bufio"
	"strings" // only needed below for sample processing
)

type user struct {
	nick, pass, name string
	chans []string
	mes chan string
	act chan string
	active bool
}

var (
	users = make(map[string]user)
	ln net.Listener
)

func cons() {
	for {
		con, _ := ln.Accept()
		// accept connection on port
		//fmt.Println("New connection)")
		go login(con)
	}

}

func ch_nick(st string) *user {
	for _, us := range users {
		if st == us.nick {
			return &us
		}
	}
	return nil
}

func ch_user(u string) bool {
	_, inc := users[u]
	return inc
}

func login(con net.Conn) {
	var n, p, u, nm string
	sc := bufio.NewScanner(con)
	for sc.Scan() {
		st := strings.Split(sc.Text(), " ")
		switch strings.ToLower(st[0]) {
		case "nick":
			if len(st) == 2 && ch_nick(st[1]) == nil {
				n = st[1]
				fmt.Println("It's nick :", n)
			} else {
				fmt.Println("Wrong nick D:")
			}
		case "user":
			if len(st) > 4 && st[4][0] == ":"[0] && !ch_user(st[1]) {
				u = st[1]
				nm = strings.Join(st[4:], " ")
				fmt.Println("It's user:", u)
			} else {
				fmt.Println("Wrong user D:")
			}
			default : fmt.Println("Nor user or nick :(  : ", st)
		}
		if len(n) > 0 && len(u) > 0 {
			break
		}
	}
	if len(n) > 0 && len(u) > 0 {
		users[u] = user{nick: n, pass: p, name: nm, act: make(chan string, 10), mes: make(chan string, 10), active: true}
		con.Write([]byte(fmt.Sprintln(":server 001 ", u, ": Hehey you're welcome")))
		go hand(u, sc, con)
	}
}

func inp(u string, sc *bufio.Scanner) {
	for sc.Scan() {
		txt := sc.Text()
		if us := ch_nick(strings.Split(txt, " ")[0]); us != nil {
			us.mes <- strings.Join(strings.Split(txt, " ")[1:], " ") + "\n"
		} else {
			fmt.Println(users[u].nick, "wrote:", sc.Text())
		}
	}
	users[u].act <- "disconnected"
}

func hand(u string, sc *bufio.Scanner, con net.Conn) {
	go inp(u, sc)
	f:
	for {
		select {
		case msg := <- users[u].mes:
			con.Write([]byte(msg))
		case act := <- users[u].act:
			switch act {
			case "disconnected":
				users[u].active = false
				break f
			}
		}
	}
	con.Close()
	fmt.Println(u, " disconnected MWAHAHA")
}

func main() {

	fmt.Println("Launching server...")

	// listen on all interfaces
	ln, _ = net.Listen("tcp", ":8081")

	go cons()
	for {
	}
}
