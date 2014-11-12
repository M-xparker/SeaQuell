package main

import (
	"fmt"
	"github.com/MattParker89/seaquell/btree"
	"github.com/MattParker89/seaquell/storage"
)

func main() {
	fmt.Println("start")
	var tree *btree.BTree
	s, err := storage.Open("test.db")
	if err != nil {
		fmt.Println("error opening db ", err)
		s, _ = storage.Create("test.db")
		tree = btree.New()
	} else {
		tree = btree.Fetch(0)
	}
	defer s.Close()

	var instruction string
	var key int
	var value int
	tree.Print()
	fmt.Println(" ")
	for {
		fmt.Scanln(&instruction, &key, &value)

		switch instruction {
		case "add":
			tree.Insert(key, []byte{byte(value)})
		case "get":
			fmt.Println("Value: ", tree.Get(key))
		}
		tree.Print()
		fmt.Println(" ")
	}
}
