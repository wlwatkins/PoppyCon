package main

import (
    // "fmt"
    // "time"
    "gobot.io/x/gobot"
    "gobot.io/x/gobot/drivers/gpio"
    "gobot.io/x/gobot/platforms/raspi"
)

type relay struct {
  Channel chan int
  Driver1 *gpio.RelayDriver
  Driver2 *gpio.RelayDriver
}

func toggleState(state *bool) {
  switch {
  case *state == false:
    *state = true
  case *state == true:
    *state = false
  }
}

func waterPump(r relay) {

  for {
        select {
        case msg := <-r.Channel:

            if msg == 1 {
              r.Driver1.Toggle()
              toggleState(&relay1State)
            } else if msg == 2 {
              r.Driver2.Toggle()
              toggleState(&relay2State)
            }
            // fmt.Println("received", msg, relay1State)
        }
  }

}

func PumpControl() *gobot.Robot {
  rpi := raspi.NewAdaptor()

  // Initiating pump relays (GPIO)
  relay1State, relay2State = false, false
  relays := relay{Channel: pumpChan,
                  Driver1: gpio.NewRelayDriver(rpi, "40"),
                  Driver2: gpio.NewRelayDriver(rpi, "38")}

  work := func() {

          relays.Driver1.On()
          relays.Driver2.On()

          relays.Driver1.Off()
          relays.Driver2.Off()
          go waterPump(relays)

          // gobot.Every(1*time.Second, func() {
          //         // fmt.Println("")
          //         pumpChan <- 1
          //         pumpChan <- 2
          //
          // })
  }

  robot := gobot.NewRobot("waterPump",
          []gobot.Connection{rpi},
          []gobot.Device{relays.Driver1, relays.Driver2},
          work,
  )

  return robot

}
