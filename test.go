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
	for {
		var n, p, u, nm string
		sc := bufio.NewScanner(con)
		for sc.Scan() {
			if sr  := sc.Text(); len(sr) < 511 {
				st := strings.Split(sr, " ");
				switch strings.ToLower(st[0]) {
				case "pass":
					if len(st) == 2 {
						p = st[1]
						fmt.Println("It's pass:", p)
					} else if len(st) == 1 {
						con.Write([]byte(fmt.Sprintln(":server 461",  strings.ToUpper(st[0]), ":Not enough parametrs")))
					} else {
						fmt.Println("Wrong pass D:")
					}
				case "nick":
					if len(st) == 2 && len(st[1]) < 10 {
						n = st[1]
						fmt.Println("It's nick :", n)
					} else {
						con.Write([]byte(fmt.Sprintln(":nickserv@service. NOTICE", st[1], ":This nickname is registered. Please choose a different nickname, or identify via /msg nickserv identify <password>.")))
						fmt.Println("Wrong nick D:")
					}
						con.Write([]byte(fmt.Sprintln(":server 461",  strings.ToUpper(st[0]), ":Not enough parametrs")))
				case "user":
					if len(st) > 4 && st[4][0] == ":"[0]{
						u = st[1]
						nm = strings.TrimPrefix(strings.Join(st[4:], " "), ":")
						fmt.Println("It's user:", u)
					} else {
						fmt.Println("Wrong user D:")
					}
					default : con.Write([]byte(fmt.Sprintln(":server 451 :You have not registered")))
				}
				if len(n) > 0 && len(u) > 0 {
					break
				}
			}
		}
		us := &user{nick: n, pass: p, name: nm, act: make(chan string, 10), mes: make(chan string, 10)}
		if (search(n, p, u, nm)) {
			_, ex := users[u]
			if !ex {
				users[u] = us
			}
			con.Write([]byte(fmt.Sprintln(":server 001 ", u, ": Hehey you're welcome", u)))
			users[u].active = true
			go hand(u, sc, con)
			break
		} else {
			con.Write([]byte(fmt.Sprintln("There is such a user and ur data doesn't match his one. Try again")))
		}

	}
}

func inp(u string, sc *bufio.Scanner, con net.Conn) {
	for sc.Scan() {
		if sr := sc.Text(); len(sr) < 511 {
			txt := strings.Split(sr, " ")
		/*	if sd := ch_nick(txt[2]); len(txt[2]) < 10 && len(txt) == 3 && sd != nil && txt[0][0] == ':' && &txt[0][1] == users[u].nick && txt[1] == "NICK" {
				users[u].nick = txt[2]
			} else if len(txt) == 1 {
				con.Write([]byte(fmt.Sprintln(":server 431", ":No nickname given")))
			} else if sd == nil {
				con.Write([]byte(fmt.Sprintln(":server 433", u, ":Nickname is alredy in use")))
			} else if sd != nil && len(txt) == 3 && len(txt[2]) > 10 && txt[0][0] == ':' && &txt[0][1] == users[u].nick && txt[1] == "NICK" {
				con.Write([]byte(fmt.Sprintln(":server 432", u, ":Errorneus nickname")))
			}
				*/
			switch strings.ToLower(txt[0]) {
			case "nick":
				if sd := ch_nick(txt[1]); len(txt) == 2 && sd != nil {
					users[u].nick = txt[1]
				}
			case "join":
			case "part":
			case "who":
			case "names":
				for  _, nk := range users {
					if nk.active == true {
						con.Write([]byte(fmt.Sprintln(":server", nk.nick, " = \"*\"")))
					}
				}
			case "list":
			case "privmsg":
				if us := ch_nick(txt[1]); us != nil && txt[2][0] == ":"[0] && len(txt[2]) > 1 {
					us.mes <- strings.TrimPrefix(strings.Join(txt[2:], " ") + "\n", ":")
				} else if us != nil && txt[2][0] == ":"[0] && len(txt[2]) == 1 {
					con.Write([]byte(fmt.Sprintln(":server 412", u, ":No text to send")))
				} else {
					fmt.Println(users[u].nick, "wrote:", sc.Text())
				}
				default: fmt.Println(users[u].nick, "wrote:", sc.Text())
			}
		}
	}
	users[u].act <- "disconnected"
}

func hand(u string, sc *bufio.Scanner, con net.Conn) {
	go inp(u, sc, con)
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
	ln, _ := net.Listen("tcp", ":8081")

	go cons(ln)
	for {
	}
}
