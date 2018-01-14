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
	active bool
	mes chan string
	act chan string
}

var (
	users = make(map[string]*user)
	chs [][]string
)

func cons(ln net.Listener) {
	for {
		con, _ := ln.Accept()
		fmt.Println("New connection)")
		go login(con)
	}
}

func ch_nick(st string) *user {
	for _, us := range users {
		if st == us.nick {
			return us
		}
	}
	return nil
}

func ch_user(u string) bool {
	_, inc := users[u]
	return inc
}

func search(n, p, u, nm string) bool {
	user, is := users[u]
	if !is {
		return true
	}
	if n == user.nick && p == user.pass && nm == user.name {
		return true
	}
	return false
}

func login(con net.Conn) {
	var n, p, u, nm string
	sc := bufio.NewScanner(con)
	for sc.Scan() {
		st := strings.Split(sc.Text(), " ")
		switch strings.ToLower(st[0]) {
		case "pass":
			if len(st) == 2 {
				p = st[1]
				fmt.Println("It's pass:", p)
			} else {
				fmt.Println("Wrong pass D:")
			}
		case "nick":
			if len(st) == 2 {
				n = st[1]
				fmt.Println("It's nick :", n)
			} else {
				fmt.Println("Wrong nick D:")
			}
		case "user":
			if len(st) > 4 && st[4][0] == ":"[0]{
				u = st[1]
				nm = strings.TrimPrefix(strings.Join(st[4:], " "), ":")
				fmt.Println("It's user:", u)
			} else {
				fmt.Println("Wrong user D:")
			}
			default : fmt.Println("Nor user, pass or nick :(  : ", st)
		}
		if len(n) > 0 && len(u) > 0 {
			us := &user{nick: n, pass: p, name: nm, act: make(chan string, 10), mes: make(chan string, 10)}
			if (search(n, p, u, nm)) {
				_, ex := users[u]
				if !ex {
					users[u] = us
					con.Write([]byte(fmt.Sprintln(":server 001 ", u, ": Hehey you're welcome")))
				}
				users[u].active = true
				go hand(u, sc, con)
				break
			} else {
				con.Write([]byte(fmt.Sprintln("There is such a user and ur data doesn't match his one. Try again")))
			}
		}
	}
}

func inp(u string, sc *bufio.Scanner) {
	for sc.Scan() {
		txt := strings.Split(sc.Text(), " ")
		switch strings.ToLower(txt[0]) {
		case "nick":
		case "join":
		case "part":
		case "who":
		case "names":
		case "list":
		case "privmsg":
			if us := ch_nick(txt[1]); us != nil && txt[2][0] == ":"[0] {
				us.mes <- strings.TrimPrefix(strings.Join(txt[2:], " ") + "\n", ":")
			} else {
				fmt.Println(users[u].nick, "wrote:", sc.Text())
			}
			default: fmt.Println(users[u].nick, "wrote:", sc.Text())
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

	go cons(ln)
	for {
	}
}
