package main

import (
	"log"

	"github.com/stapelberg/emoji-picker-for-i3/internal/picker"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	if err := picker.Run(); err != nil {
		log.Fatal(err)
	}
}
