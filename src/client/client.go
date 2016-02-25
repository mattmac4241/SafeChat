package main

import (
	"bufio"
	"crypt"
	"crypto/rsa"
	"errors"
	"fmt"
	"helper"
	"log"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

var (
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	userKey    *rsa.PublicKey
	connected  bool
	logged_in  bool
	connection net.Conn
)

func main() {
	listener, err := createListener(os.Args)
	if err != nil {
		fmt.Println(err)
	} else {
		connection = listener
		privateKey = crypt.GenPrivateKey()
		publicKey = crypt.GetPublicKey(privateKey)
		go handleInput()      // handle input
		go handleConnection() // handle connections
		for {

		}
		listener.Close()
	}
}

func handleConnection() {
	for {
		mess := helper.DecodeMessage(connection)
		evalMessage(mess)
	}
}

func handleInput() {
	fmt.Println("Log in or Register")
	input := bufio.NewScanner(os.Stdin)
	if logged_in == true {
		for input.Scan() {
			text := input.Text()
			helper.EncodeMessage(*evalInput(text), connection)
		}
	} else {
		loginOrRegister()
	}
}

func evalInput(text string) *helper.Message {
	var message *helper.Message
	isCommand := isCommand(text)
	if isCommand == true {
		message = helper.CreateMessage(nil, text)
	} else {
		if connected == true && userKey != nil {
			mess := crypt.Encrypt(userKey, []byte(text))
			message = helper.CreateMessage(mess, "Encrypted")
		} else {
			message = helper.CreateMessage([]byte(text), "Message")
		}
	}
	return message
}

func isCommand(text string) bool {
	if strings.HasPrefix(text, "-c") {
		return true
	} else {
		switch text {
		case "-h", "-g", "-d":
			return true
		default:
			return false

		}
	}

}

func evalMessage(message helper.Message) {
	switch message.Command {
	case "Server", "Message":
		serverMessage(string(message.Message))
	case "SendKey":
		fmt.Println("Now connected")
		sendKey(message)
	case "Key":
		key(message)
	case "Encrypted":
		decryptMessage(message)
	default:
		fmt.Println("Not valid message")
	}
}

func decryptMessage(message helper.Message) {
	mess := crypt.Decrypt(privateKey, message.Message)
	result := fmt.Sprintf("%s: %s", message.From, string(mess))
	fmt.Println(result)
}

func serverMessage(message string) {
	fmt.Println(message)
}

func sendKey(message helper.Message) {
	connected = true
	mess := helper.CreateMessage(nil, "Key")
	mess.Key = publicKey
	helper.EncodeMessage(*mess, connection)
}

func key(message helper.Message) {
	if message.Key == nil {
		fmt.Println("Key not sent")
	} else {
		userKey = message.Key
	}
}

func createListener(args []string) (net.Conn, error) {
	if len(args[1:]) != 2 {
		return nil, errors.New("Server and Port required")
	} else {
		server := args[1] // get the server to connect to
		port := args[2]   // get the port number
		conn := fmt.Sprintf("%s:%s", server, port)
		listener, err := net.Dial("tcp", conn)
		if err != nil {
			log.Fatal(err)
		}
		return listener, nil
	}
}

func loginOrRegister() {
	reader := bufio.NewReader(os.Stdin)
	command, _ := reader.ReadString('\n')
	command = strings.ToLower(command)
	command = strings.TrimSpace(command)
	if command == "login" || command == "register" {
		fmt.Print("Enter username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)
		fmt.Print("Enter password: ")
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)
		if command == "login" {
			login(username, password)
		} else {
			register(username, password)
		}

	} else {
		fmt.Println("Not valid")
		loginOrRegister()
	}
}

func login(username, password string) {
	cred := helper.Credential{username, []byte(password)}
	fmt.Println(cred)
	data, err := bson.Marshal(cred)
	if err != nil {
		panic(err)
	}
	message := helper.CreateMessage(data, "Login")
	helper.EncodeMessage(*message, connection)
}

func register(username, password string) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	cred := helper.Credential{username, hashed}
	data, err := bson.Marshal(cred)
	if err != nil {
		panic(err)
	}
	message := helper.CreateMessage(data, "Register")
	helper.EncodeMessage(*message, connection)
}
