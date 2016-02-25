package main

import (
	dbm "database_manager"
	"fmt"
	"helper"
	"log"
	"net"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	dbCol     *mgo.Collection
)

func main() {
	userIndex = 0
	listener, err := net.Listen("tcp", "localhost:5000")
	if err != nil {
		log.Fatal(err)
	}
	users = make(map[string]*User)

	session, err := mgo.Dial("mongodb://matt:home222@127.0.01:27017/test")
	defer session.Close()
	if err != nil {
		panic(err)
	}
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	dbCol = session.DB("test").C("users")

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
		evalMessage(message, user, c)
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

func evalMessage(message helper.Message, user *User, c net.Conn) {
	message.From = user.username
	if strings.HasPrefix(message.Command, "-c") {
		connectTo(user, message.Command)
	} else {
		switch message.Command {
		case "Message", "Encrypted":
			userMessage(message, user)
		case "-g":
			getUsers(user)
		case "-d":
			disconnect(user)
		case "Key":
			userMessage(message, user)
		case "Login":
			login(message, c)
		case "Register":
			register(message, c)
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
	userDisconnect(user.connectedTo)
	userDisconnect(user)
}

func userDisconnect(user *User) {
	if user.isConnected {
		message := helper.CreateMessage([]byte("Disconnected from user"), "Server")
		sendMessage(*message, user)
		user.connectedTo = nil
		user.isConnected = false
	} else {
		message := helper.CreateMessage([]byte("Not connected to anyone"), "Server")
		sendMessage(*message, user)
	}
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

func login(message helper.Message, c net.Conn) {
	cred := new(helper.Credential)
	bson.Unmarshal(message.Message, cred)
	fmt.Println(cred)
}

func register(message helper.Message, c net.Conn) {
	cred := new(helper.Credential)
	bson.Unmarshal(message.Message, cred)
	if dbm.IsUser(cred.Username, dbCol) {
		mess := helper.CreateMessage([]byte("Username already taken"), "Server")
		helper.EncodeMessage(*mess, c)
	} else {
		dbm.InsertUser(cred.Username, string(cred.Password), dbCol)
		mess := helper.CreateMessage([]byte("Registered"), "Register")
		helper.EncodeMessage(*mess, c)
	}
}
