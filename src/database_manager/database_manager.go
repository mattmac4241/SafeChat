package database_manager

import (
	"errors"
	"log"

	"gopkg.in/mgo.v2"

	"gopkg.in/mgo.v2/bson"
)

type User struct {
	UserName string
	Password string
}

//insertUser into database
func InsertUser(username, password string, c *mgo.Collection) {
	err := c.Insert(&User{username, password})
	if err != nil {
		log.Fatal(err)
	}
}

//check if user already exists
func IsUser(username string, c *mgo.Collection) bool {
	result := User{}
	c.Find(bson.M{"username": username}).One(&result)
	if result == (User{}) {
		return false
	} else {
		return true
	}
}

func GetUser(username string, c *mgo.Collection) (User, error) {
	result := User{}
	c.Find(bson.M{"username": username}).One(&result)
	if result == (User{}) {
		return User{}, errors.New("User not found")
	} else {
		return result, nil
	}
}
