package main

import (
	"fmt"
	"helper"
	"log"
	"net"
	"strings"
)

type User struct {
	isConnected bool
	connection  net.Conn
	connectedTo *User
	username    string
}

var (
	users     map[string]*User
	userIndex int
)

func main() {
	userIndex = 0
	listener, err := net.Listen("tcp", "localhost:5000")
	if err != nil {
		log.Fatal(err)
	}
	users = make(map[string]*User)

	for {
		userIndex += 1
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(c net.Conn) {
	userName := fmt.Sprintf("user%d", userIndex)
	user := createUser(false, c, userName)
	users[userName] = user
	for {
		message := helper.DecodeMessage(c)
		if message.Command == "" {
			break
		}
		evalMessage(message, user)
	}
	c.Close()

}

func createUser(isConnected bool, c net.Conn, username string) *User {
	user := new(User)
	user.isConnected = isConnected
	user.connection = c
	user.username = username
	return user
}

func evalMessage(message helper.Message, user *User) {
	if strings.HasPrefix(message.Command, "-c") {
		connectTo(user, message.Command)
	} else {
		switch message.Command {
		case "Message", "Encrypted":
			message.From = user.username
			userMessage(message, user)
		case "-g":
			getUsers(user)

		case "-d":
			disconnect(user)
		case "Key":
			userMessage(message, user)
		default:
			message := helper.CreateMessage([]byte("Not a valid command"), "Server")
			sendMessage(*message, user)
		}
	}
}

func userMessage(message helper.Message, user *User) {
	if user.isConnected {
		sendMessage(message, user.connectedTo)
	} else {
		mess := helper.CreateMessage([]byte("Not connected to another user"), "Server")
		sendMessage(*mess, user)
	}
}

func getUsers(user *User) {
	userList := ""
	for userName, _ := range users {
		if userName == user.username {
			continue
		}
		userList += fmt.Sprintf("%s\n", userName)
	}
	if userList == "" {
		userList = "No one is connected"
	}
	message := helper.CreateMessage([]byte(userList), "Server")
	sendMessage(*message, user)
}

func disconnect(user *User) {
	userTo := user.connectedTo.connectedTo
	mess := fmt.Sprintf("%s left", user.username)
	message := helper.CreateMessage([]byte(mess), "Server")
	sendMessage(*message, userTo)
	user.connectedTo.connectedTo = nil
	user.connectedTo = nil
	user.isConnected = false
	userTo.isConnected = false
}

func connectTo(user *User, message string) {
	parts := strings.Split(message, " ")
	if len(parts) != 2 {
		mess := helper.CreateMessage([]byte("Not a valid Connection"), "Server")
		sendMessage(*mess, user)
	} else if isUser(parts[1]) {
		userTo := users[parts[1]]
		if userTo.isConnected {
			mess := helper.CreateMessage([]byte("User already connected"), "Server")
			sendMessage(*mess, user)
		} else {
			connectUsers(user, userTo)
		}
	} else {
		mess := helper.CreateMessage([]byte("Not a valid user"), "Server")
		sendMessage(*mess, user)
	}
}

func sendMessage(message helper.Message, user *User) {
	conn := user.connection
	helper.EncodeMessage(message, conn)

}

func isUser(name string) bool {
	if _, userTo := users[name]; userTo {
		return true
	} else {
		return false
	}
}

func messageMultiUsers(message helper.Message, users []*User) {
	for _, user := range users {
		sendMessage(message, user)
	}
}

func connectUsers(user1, user2 *User) {
	user1.isConnected = true
	user1.connectedTo = user2
	user2.isConnected = true
	user2.connectedTo = user1
	userSlice := []*User{user1, user2}
	keyMessage := helper.CreateMessage([]byte("Send Key"), "SendKey")
	messageMultiUsers(*keyMessage, userSlice)

}
