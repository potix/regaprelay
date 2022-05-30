package main

import (
	"os"
	"log"
	"encoding/hex"
)

var checkAddr60 = []string{
	"6000",
	"6010",
	"6020",
	"6030",
	"6040",
	"6050",
	"6060",
	"6070",
	"6080",
	"6090",
	"60a0",
}

var checkAddr80 = []string{
	"8000",
	"8010",
	"8020",
	"8030",
}

func main() {
	rw, err := os.OpenFile("/dev/hidraw0", os.O_RDWR, 0644)
	if err != nil {
                log.Printf("can not open /dev/hidraw: %v", err)
		return
	}
	defer rw.Close()

	macaddr := make([]byte, 6)

	buf := make([]byte, 128)
	_, err = rw.Write([]byte{ 0x80, 0x01 })
	if err != nil {
		log.Printf("can not write: %v", err)
		return
	}
	for {
		rl, err := rw.Read(buf)
		if err != nil {
			log.Printf("can not read: %v", err)
			return
		}
		if buf[0] == 0x30 {
			continue
		}
		if rl < 10 {
			log.Printf("can not get mac address (%v): %v", buf[:rl], err)
			return
		}
		log.Printf("=====================")
		log.Printf("reponse code %x", buf[:2])
		log.Printf("padding %x", buf[2:3])
		log.Printf("device type %x", buf[3:4])
		log.Printf("macaddress %x", buf[4:10])
		var i = 0
		for _, b := range buf[4:10] {
			macaddr[i] = b
			i += 1
		}
		break
	}


	_, err = rw.Write([]byte{ 0x80, 0x02 })
	if err != nil {
		log.Printf("can not write: %v", err)
		return
	}
	for {
		rl, err := rw.Read(buf)
		if err != nil {
			log.Printf("can not read: %v", err)
			return
		}
		if buf[0] == 0x30 {
			continue
		}
		if rl < 2 {
			log.Printf("can not handshake (%v): %v", buf[:rl], err)
			return
		}
		log.Printf("=====================")
		log.Printf("reponse code %x", buf[:2])
		break
	}

	_, err = rw.Write([]byte{ 0x80, 0x04 })
	if err != nil {
		log.Printf("can not write: %v", err)
		return
	}

	memory60 := make([]byte, 0, 200)
	memory80 := make([]byte, 0, 100)

	var counter byte = 0
	for _, addr := range checkAddr60 {
		addrBytes, err := hex.DecodeString(addr)
		if err != nil {
			log.Printf("can not decode string: %v", err)
			return
		}
		counter += 1
		if err != nil {
			log.Printf("can not write: %v", err)
		}
		log.Printf("write %x", []byte{ 0x01, counter, 0, 0, 0, 0, 0, 0, 0, 0, 0x10, addrBytes[1], addrBytes[0], 0, 0, 0x10 })
		_, err = rw.Write([]byte{ 0x01, counter, 0, 0, 0, 0, 0, 0, 0, 0, 0x10, addrBytes[1], addrBytes[0], 0, 0, 0x10 })
		if err != nil {
			log.Printf("can not write: %v", err)
			return
		}

		for {
			rl, err := rw.Read(buf)
			if err != nil {
				log.Printf("can not read: %v", err)
				return
			}
			if buf[0] == 0x30 {
				continue
			}
			if rl < 37 {
				log.Printf("can not get spi memory (%v): %v", buf[:rl], err)
				return
			}
			log.Printf("=====================")
			log.Printf("entire %x", buf[:rl])
			log.Printf("report id %x", buf[0])
			log.Printf("controller data %x", buf[1:13])
			log.Printf("ack %x", buf[13:14])
			log.Printf("subcommand %x", buf[14:15])
			log.Printf("addr %x%x", buf[16:17], buf[15:16])
			log.Printf("paddinf %x", buf[17:19])
			log.Printf("length %x", buf[19:20])
			log.Printf("memory data %x", buf[20:36])
			memory60 = append(memory60, buf[20:36]...)
			break
		}

	}

	for _, addr := range checkAddr80 {
		addrBytes, err := hex.DecodeString(addr)
		if err != nil {
			log.Printf("can not decode string: %v", err)
			return
		}
		counter += 1
		if err != nil {
			log.Printf("can not write: %v", err)
		}
		log.Printf("write %x", []byte{ 0x01, counter, 0, 0, 0, 0, 0, 0, 0, 0, 0x10, addrBytes[1], addrBytes[0], 0, 0, 0x10 })
		_, err = rw.Write([]byte{ 0x01, counter, 0, 0, 0, 0, 0, 0, 0, 0, 0x10, addrBytes[1], addrBytes[0], 0, 0, 0x10 })
		if err != nil {
			log.Printf("can not write: %v", err)
			return
		}

		for {
			rl, err := rw.Read(buf)
			if err != nil {
				log.Printf("can not read: %v", err)
				return
			}
			if buf[0] == 0x30 {
				continue
			}
			if rl < 37 {
				log.Printf("can not get spi memory (%v): %v", buf[:rl], err)
				return
			}
			log.Printf("=====================")
			log.Printf("entire %x", buf[:rl])
			log.Printf("report id %x", buf[0])
			log.Printf("controller data %x", buf[1:13])
			log.Printf("ack %x", buf[13:14])
			log.Printf("subcommand %x", buf[14:15])
			log.Printf("addr %x%x", buf[16:17], buf[15:16])
			log.Printf("paddinf %x", buf[17:19])
			log.Printf("length %x", buf[19:20])
			log.Printf("memory data %x", buf[20:36])
			memory80 = append(memory80, buf[20:36]...)
			break
		}
	}
	log.Printf("=====================")
	log.Printf("macaddr %x", macaddr)
	log.Printf("spiMemoryDump60 %x", memory60)
	log.Printf("spiMemoryDump80 %x", memory80)

}
