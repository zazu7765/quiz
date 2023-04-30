package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
	"os"
)

func createCardDeck(db *sql.DB, name string) {
}
func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(home + "/.quiz"); os.IsNotExist(err) {
		fmt.Println("make directory")
		os.Mkdir(home+"/.quiz", 0750)
	} else {
		fmt.Println("directory already exists")
	}
}
