package main

import (
	"fmt"
	"log"

	"github.com/MichaelS11/go-dht"
)

func main() {
	err := dht.HostInit()
	if err != nil {
		log.Println("HostInit error:", err)
		return
	}

	dht, err := dht.NewDHT("GPIO26", dht.Celsius, "")
	if err != nil {
		log.Println("NewDHT error:", err)
		return
	}
	for {
		humidity, temperature, err := dht.Read()
		if err != nil {
			log.Println("Read error:", err)
			continue
		}

		fmt.Printf("humidity: %v\n", humidity)
		fmt.Printf("temperature: %v\n", temperature)
	}
}
