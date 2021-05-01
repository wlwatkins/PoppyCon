package main

import (
	"fmt"

	"github.com/MichaelS11/go-dht"
)

func main() {
	err := dht.HostInit()
	if err != nil {
		fmt.Println("HostInit error:", err)
		return
	}

	dht, err := dht.NewDHT("GPIO26", dht.Celsius, "")
	if err != nil {
		fmt.Println("NewDHT error:", err)
		return
	}
  for {
  	humidity, temperature, err := dht.Read()
  	if err != nil {
  		fmt.Println("Read error:", err)
  		continue
  	}

  	fmt.Printf("humidity: %v\n", humidity)
  	fmt.Printf("temperature: %v\n", temperature)
  }
}
