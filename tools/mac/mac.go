package main

import (
	"os"
	"log"
)

func main() {
	rw, err := os.OpenFile("/dev/hidraw0", os.O_RDWR, 0644)
	if err != nil {
                log.Printf("can not open /dev/hidraw: %v", err)
		return
	}
	defer rw.Close()

	buf := make([]byte, 1024)
	_, err = rw.Write([]byte{ 0x80, 0x01 })
	if err != nil {
		log.Printf("can not write: %v", err)
	}
	rl, err := rw.Read(buf)
	if err != nil {
		log.Printf("can not read: %v", err)
		return
	}
	if rl < 10 {
		log.Printf("can not get mac address (%v): %v", buf[:rl], err)
		return
	} else {
		log.Printf("reponse code %x", buf[:2])
		log.Printf("padding %x", buf[2:3])
		log.Printf("device type %x", buf[3:4])
		log.Printf("macaddress %x", buf[4:10])
	}
	
}
