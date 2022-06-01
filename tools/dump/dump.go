package main

import (
	"os"
	"log"
	"syscall"
	"github.com/potix/utils/signal"
	"github.com/potix/regaprelay/gamepad"
	"errors"
)

func rwLoop(rf *os.File, wf *os.File, direction string) {
	buf := make([]byte, 1024)
	for {
		rl, err := rf.Read(buf)
		if err != nil {
			log.Printf("can not read (%v): %v", direction, err)
			return
		}
		for {
			_, err := wf.Write(buf[:rl])
			if err != nil {

				switch {
				case errors.Is(err, syscall.EAGAIN):
					continue
				default:
					log.Printf("can not write (%v): %v", direction, err)
					return
				}
			}
			break
		}
		log.Printf("%v: %x", direction, buf[:rl])
	}
}

func main() {
        _, err := gamepad.NewGamepad(gamepad.ModelNSProCon, "", "", "")
        if err != nil {
                log.Fatalf("can not create gamepad: %v", err)
        }
	fhidg0, err := os.OpenFile("/dev/hidg0", os.O_RDWR, 0644)
	if err != nil {
                log.Printf("can not open /dev/hidg0: %v", err)
		return
	}
	defer fhidg0.Close()
	fhidraw, err := os.OpenFile("/dev/hidraw0", os.O_RDWR, 0644)
	if err != nil {
                log.Printf("can not open /dev/hidraw: %v", err)
		return
	}
	defer fhidraw.Close()
	go rwLoop(fhidg0, fhidraw, "switch -> procon")
	go rwLoop(fhidraw, fhidg0, "procon -> switch")
	signal.SignalWait(nil)
}
