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
)

var (
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	userKey    *rsa.PublicKey
	connected  bool
)

func main() {
	listener, err := createListener(os.Args)
	if err != nil {
		fmt.Println(err)
	} else {
		privateKey = crypt.GenPrivateKey()
		publicKey = crypt.GetPublicKey(privateKey)
		go handleInput(listener)      // handle input
		go handleConnection(listener) // handle connections
		for {

		}
		listener.Close()
	}

}

func handleConnection(c net.Conn) {
	for {
		mess := helper.DecodeMessage(c)
		evalMessage(mess, c)
	}
}

func handleInput(c net.Conn) {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		text := input.Text()
		helper.EncodeMessage(*evalInput(text), c)
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

func evalMessage(message helper.Message, c net.Conn) {
	switch message.Command {
	case "Server", "Message":
		serverMessage(string(message.Message))
	case "SendKey":
		fmt.Println("Now connected")
		sendKey(message, c)
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

func sendKey(message helper.Message, c net.Conn) {
	connected = true
	mess := helper.CreateMessage(nil, "Key")
	mess.Key = publicKey
	helper.EncodeMessage(*mess, c)
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
		connection := fmt.Sprintf("%s:%s", server, port)
		listener, err := net.Dial("tcp", connection)
		if err != nil {
			log.Fatal(err)
		}
		return listener, nil
	}
}
