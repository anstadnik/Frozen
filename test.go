package main

import (
	"net"
	"fmt"
	"bufio"
	"strings" // only needed below for sample processing
)

type user struct {
	nick, pass, name string
	chans map[string]bool
	active bool
	mes chan string
	act chan string
}

type ch struct {
	mes chan string
	act chan string
	users map[string]bool
}

var (
	users = make(map[string]*user)
	chs = make(map[string]*ch)
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
			if sr := sc.Text(); len(sr) < 511 {
				st := strings.Split(sr, " ")
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
					if len(st) == 2 && len(st[1]) < 10{
						n = st[1]
						fmt.Println("It's nick :", n)
					} else if len(st) == 1 {
						fmt.Println("Wrong nick D:")
						con.Write([]byte(fmt.Sprintln(":server 461",  strings.ToUpper(st[0]), ":Not enough parametrs")))
					} else if len(st) == 2 && len(st[1]) > 10 {
						con.Write([]byte(fmt.Sprintln(":server 432", u, ":Errorneus nickname")))
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
						con.Write([]byte(fmt.Sprintln(":server 451 :You have not registered")))
				}
				if len(n) > 0 && len(u) > 0 {
					break
				}
			}
		}
		us := &user{nick: n, pass: p, name: nm, act: make(chan string, 10), mes: make(chan string, 100), chans: make(map[string]bool)}
		if (search(n, p, u, nm)) {
			_, ex := users[u]
			if !ex && ch_nick(n) != nil {
				con.Write([]byte(fmt.Sprintln(":nickserv@service. NOTICE", n, ":This nickname is registered. Please choose a different nickname, or identify via /msg nickserv identify <password>.")))
				continue
			}
			if !ex {
				users[u] = us
			}
			con.Write([]byte(fmt.Sprintln(":server 001 ", u, ": Hehey you're welcome")))
			users[u].active = true
			go hand(u, sc, con)
			break
		} else {
			con.Write([]byte(fmt.Sprintln(":nickserv@service. NOTICE", n, ":This nickname is registered. Please choose a different nickname, or identify via /msg nickserv identify <password>.")))
		}

	}
}

func ch_handl(n string) {
	//fmt.Println("I am ch_handl")
	for {
		select {
		case msg := <- chs[n].mes:
			for u, _ := range chs[n].users {
				users[u].mes <- msg
			}
		case act := <- chs[n].act:
			words := strings.Split(act, " ")
			switch words[0] {
			case "new":
				chs[n].users[words[1]] = true
				chs[n].mes <- "New user :" + words[1] + "\n"
			case "left":
				chs[n].mes <- words[1] + "left((((" + "\n"
			}
		}
	}
}

func ch_ex(s string) bool {
	//fmt.Println("I am ch_ex")
	_, ex := chs[s]
	return ex
}

func uch_ex(u, s string) bool {
	//fmt.Println("I am uch_ex")
	_, ex := users[u].chans[s]
	return ex
}

func inp(u string, sc *bufio.Scanner, con net.Conn) {
	for sc.Scan() {
		if sr := sc.Text(); len(sr) < 511 {
			txt := strings.Split(sr, " ")
			switch strings.ToLower(txt[0]) {
			case "join":
			//fmt.Println("join")
				if len(txt) == 2 {
					users[u].chans[txt[1]] = true
					c, ex := chs[txt[1]]
					if ex {
						//fmt.Println("Adding to the channel")
						c.act <- "new " + u
					} else {
						//fmt.Println("Creating a new channel")
						chs[txt[1]] = &ch{act: make(chan string, 10), mes: make(chan string, 10), users: make(map[string]bool)}
						chs[txt[1]].act <- "new " + u
						//fmt.Println("go ch_handl")
						go ch_handl(txt[1])
					}
				} else {
					con.Write([]byte("Error name of channel"))
				}
			case "part":
				//fmt.Println("part")
				if len(txt) == 2 && ch_ex(txt[1]){
					delete(users[u].chans, txt[1])
					delete(chs[txt[1]].users, u)
					chs[txt[1]].act <- "left " + u
				} else {
					con.Write([]byte("Error name of channel"))
				}
			case "who":
				//fmt.Println("who")
			case "names":
				//fmt.Println("names")
				for  _, nk := range users {
					if nk.active == true {
						con.Write([]byte(fmt.Sprintln(":server", nk.nick, " = \"*\"")))
					}
				}
			case "list":
				//fmt.Println("list")
				if len(txt) == 1 {
					for k, _ := range chs {
						con.Write([]byte(k + "\n"))
					}
				} else if len(txt) > 1 {
					for _, u := range txt[1:] {
						if ch_ex(u) {
							con.Write([]byte(u + "\n"))
						}
					}
				} else {
					con.Write([]byte("Error name of channel\n"))
				}
			case "privmsg":
				fmt.Println("privmsg")
				us := ch_nick(txt[1]);
				if  len(txt) > 2 && us != nil && txt[2][0] == ":"[0] && len(txt[2]) > 1 {
					fmt.Println("Writing to the person")
					us.mes <- strings.TrimPrefix(strings.Join(txt[2:], " ") + "\n", ":")
				} else if us != nil && txt[2][0] == ":"[0] && len(txt[2]) == 1 {
					con.Write([]byte(fmt.Sprintln(":server 412", u, ":No text to send")))
				} else if ch_ex(txt[1]) && uch_ex(u, txt[1]) {
					fmt.Println("Writing to the channel")
					chs[txt[1]].mes <- u + " wrote: " + strings.TrimPrefix(strings.Join(txt[2:], " ") + "\n", ":")
				} else {
					fmt.Println(users[u].nick, "wrote:", strings.Join(txt, " "))
				}
			default:
				if len(txt) == 3 && len(txt[2]) < 10 && ch_nick(txt[2]) == nil && txt[0][0] == ':' && txt[0][1:] == users[u].nick && strings.ToLower(txt[1]) == "nick" {
					users[u].nick = txt[2]
				} else if len(txt) == 2 && strings.ToLower(txt[1]) == "nick" {
					con.Write([]byte(fmt.Sprintln(":server 431", ":No nickname given")))
				} else if len(txt) == 3 && ch_nick(txt[2]) != nil {
					con.Write([]byte(fmt.Sprintln(":server 433", u, ":Nickname is alredy in use")))
				} else {
					con.Write([]byte(fmt.Sprintln(":server 432", u, ":Errorneus nickname")))
				}
				fmt.Println("default")
				fmt.Println(users[u].nick, "wrote:", strings.Join(txt, " "))
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
