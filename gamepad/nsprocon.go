package gamepad

import(
	"fmt"
	"log"
	"time"
	"math"
	"os"
	"github.com/potix/regaprelay/gamepad/setup"
	"github.com/potix/regapweb/message"
	"encoding/hex"
)

type comState int

const (
	comStateInit               comState = iota
	comStateEnableUsbTimeout
	comStateMac
	comStateHandshake
	comStateBaudRate
	comStateHandshake2
	comStateDisableUsbTimeout
)


// Report IDs.
const (
	reportIdInput01    byte = 0x01
	reportIdInput10         = 0x10
	reportIdOutput21        = 0x21
	reportIdOutput30        = 0x30
	usbReportIdInput80      = 0x80
	usbReportIdOutput81     = 0x81
)

// Sub-types of the 0x81 input report, used for initialization.
const (
	subTypeRequestMac        byte = 0x01
	subTypeHandshake              = 0x02
	subTypeBaudRate               = 0x03
	subTypeDisableUsbTimeout      = 0x04
	subTypeEnableUsbTimeout       = 0x05
)

// Values for the |device_type| field reported in the MAC reply.
const (
	usbDeviceTypeChargingGripNoDevice byte = 0x00
	usbDeviceTypeChargingGripJoyConL       = 0x01
	usbDeviceTypeChargingGripJoyConR       = 0x02
	usbDeviceTypeProController             = 0x03
)

// UART subcommands.
const (
	subCommandBluetoothManualPairing   byte = 0x01
	subCommandRequestDeviceInfo             = 0x02
	subCommandSetInputReportMode            = 0x03
	subCommandTriggerButtonsElapsedTime     = 0x04
	subCommandSetHciState                   = 0x06
	subCommandSetShipmentLowPowerState      = 0x08
	subCommandReadSpi                       = 0x10
	subCommandSetNfcIrMcuConfiguration      = 0x21
	subCommandSetPlayerLights               = 0x30
	subCommand33                            = 0x33
	subCommandSetHomeLight                  = 0x38
	subCommandEnableImu                     = 0x40
	subCommandSetImuSensitivity             = 0x41
	subCommandEnableVibration               = 0x48
)

var vibrationAmpHfaMap map[uint8]int = map[uint8]int{
    0x00: 0,   0x02: 10,   0x04: 12,
    0x06: 14,  0x08: 17,   0x0a: 20,
    0x0c: 24,  0x0e: 28,   0x10: 33,
    0x12: 40,  0x14: 47,   0x16: 56,
    0x18: 67,  0x1a: 80,   0x1c: 95,
    0x1e: 112, 0x20: 117,  0x22: 123,
    0x24: 128, 0x26: 134,  0x28: 140,
    0x2a: 146, 0x2c: 152,  0x2e: 159,
    0x30: 166, 0x32: 173,  0x34: 181,
    0x36: 189, 0x38: 198,  0x3a: 206,
    0x3c: 215, 0x3e: 225,  0x40: 230,
    0x42: 235, 0x44: 240,  0x46: 245,
    0x48: 251, 0x4a: 256,  0x4c: 262,
    0x4e: 268, 0x50: 273,  0x52: 279,
    0x54: 286, 0x56: 292,  0x58: 298,
    0x5a: 305, 0x5c: 311,  0x5e: 318,
    0x60: 325, 0x62: 332,  0x64: 340,
    0x66: 347, 0x68: 355,  0x6a: 362,
    0x6c: 370, 0x6e: 378,  0x70: 387,
    0x72: 395, 0x74: 404,  0x76: 413,
    0x78: 422, 0x7a: 431,  0x7c: 440,
    0x7e: 450, 0x80: 460,  0x82: 470,
    0x84: 480, 0x86: 491,  0x88: 501,
    0x8a: 512, 0x8c: 524,  0x8e: 535,
    0x90: 547, 0x92: 559,  0x94: 571,
    0x96: 584, 0x98: 596,  0x9a: 609,
    0x9c: 623, 0x9e: 636,  0xa0: 650,
    0xa2: 665, 0xa4: 679,  0xa6: 694,
    0xa8: 709, 0xaa: 725,  0xac: 741,
    0xae: 757, 0xb0: 773,  0xb2: 790,
    0xb4: 808, 0xb6: 825,  0xb8: 843,
    0xba: 862, 0xbc: 881,  0xbe: 900,
    0xc0: 920, 0xc2: 940,  0xc4: 960,
    0xc6: 981, 0xc8: 1000,
}

var vibrationAmpLfaMap map[uint16]int = map[uint16]int{
    0x0040: 0,   0x8040: 10,   0x0041: 12,
    0x8041: 14,  0x0042: 17,   0x8042: 20,
    0x0043: 24,  0x8043: 28,   0x0044: 33,
    0x8044: 40,  0x0045: 47,   0x8045: 56,
    0x0046: 67,  0x8046: 80,   0x0047: 95,
    0x8047: 112, 0x0048: 117,  0x8048: 123,
    0x0049: 128, 0x8049: 134,  0x004a: 140,
    0x804a: 146, 0x004b: 152,  0x804b: 159,
    0x004c: 166, 0x804c: 173,  0x004d: 181,
    0x804d: 189, 0x004e: 198,  0x804e: 206,
    0x004f: 215, 0x804f: 225,  0x0050: 230,
    0x8050: 235, 0x0051: 240,  0x8051: 245,
    0x0052: 251, 0x8052: 256,  0x0053: 262,
    0x8053: 268, 0x0054: 273,  0x8054: 279,
    0x0055: 286, 0x8055: 292,  0x0056: 298,
    0x8056: 305, 0x0057: 311,  0x8057: 318,
    0x0058: 325, 0x8058: 332,  0x0059: 340,
    0x8059: 347, 0x005a: 355,  0x805a: 362,
    0x005b: 370, 0x805b: 378,  0x005c: 387,
    0x805c: 395, 0x005d: 404,  0x805d: 413,
    0x005e: 422, 0x805e: 431,  0x005f: 440,
    0x805f: 450, 0x0060: 460,  0x8060: 470,
    0x0061: 480, 0x8061: 491,  0x0062: 501,
    0x8062: 512, 0x0063: 524,  0x8063: 535,
    0x0064: 547, 0x8064: 559,  0x0065: 571,
    0x8065: 584, 0x0066: 596,  0x8066: 609,
    0x0067: 623, 0x8067: 636,  0x0068: 650,
    0x8068: 665, 0x0069: 679,  0x8069: 694,
    0x006a: 709, 0x806a: 725,  0x006b: 741,
    0x806b: 757, 0x006c: 773,  0x806c: 790,
    0x006d: 808, 0x806d: 825,  0x006e: 843,
    0x806e: 862, 0x006f: 881,  0x806f: 900,
    0x0070: 920, 0x8070: 940,  0x0071: 960,
    0x8071: 981, 0x0072: 1000,
}

type buttons struct {
        a            byte
        b            byte
        x            byte
	y            byte
        l            byte
        r            byte
        zl           byte
        zr           byte
        minus        byte
        plus         byte
        home         byte
        capture      byte
        left         byte
        right        byte
        up           byte
        down         byte
        leftSl       byte
        leftSr       byte
        rightSl      byte
        rightSr      byte
        chargingGrip byte
}

type stick struct {
	x     float64
	y     float64
	press byte
}

type controller struct {
	buttons    *buttons
	leftStick  *stick
	rightStick *stick
}

type imuSensitivity struct {
	gyroSensitivity              byte
	accelerometerSensitivity     byte
	gyroPerformanceRate          byte
	accelerometerFilterBandwidth byte

}

type NSProCon struct  {
	*BaseBackend
	setupParams     *setup.UsbGadgetHidSetupParams
	verbose         bool
	macAddr         []byte
	reverseMacAddr  []byte
	spiMemory60     []byte
	spiMemory80     []byte
	devFilePath     string
	devFile         *os.File
	comState        comState
	usbTimeout      bool
	reportCounter   byte
	imuEnable       byte
	vibrationEnable byte
	stopCh          chan int
	controller      *controller
	imuSensitivity  *imuSensitivity
}

func (n *NSProCon) writeReport(f *os.File, reportId byte, reportBytes []byte) (error) {
	buf := make([]byte, 64)
	buf[0] = reportId
	for i, b := range reportBytes {
		buf[i + 1] = b
	}
	wl, err := f.Write(buf)
	if err != nil {
		return fmt.Errorf("can not write report (%x) to gadget device file: %w", reportId,  err)
	}
	if wl != len(buf) {
		return fmt.Errorf("partial write report (%x) to gadget device file: write len = %v", reportId, wl)
	}
	if n.verbose {
		log.Printf("wrote %x", buf)
	}
	return nil
}




func (n *NSProCon) sendVibrationRequest(bytes []byte) error {
	if len(bytes) < 8 {
		return fmt.Errorf("invalid vibration data (%x)", bytes)
	}
	lhamp := 0
	llamp := 0
	rhamp := 0
	rlamp := 0
	leftSkip := true
	rightSkip := true
	ok := false
	if bytes[0] == 0 && bytes[1] == 0 && bytes[2] == 0 && bytes[3] == 0 {
		leftSkip = true
	}
	if bytes[4] == 0 && bytes[5] == 0 && bytes[6] == 0 && bytes[7] == 0 {
		rightSkip = true
	}
	if !leftSkip {
		//var lhf uint16 = uint16(bytes[1]&0x01)<<8 | uint16(bytes[0])
		var lhfAmp uint8 = uint8(bytes[1] & 0xfe)
		//var llf uint8 = uint8(bytes[2] & 0x7f)
		var llfAmp uint16 = uint16(bytes[2]&0x80)<<8 | uint16(bytes[3])
		lhamp, ok = vibrationAmpHfaMap[lhfAmp]
		if !ok {
			return fmt.Errorf("ont found left hight amplitude (%v): %+v", lhfAmp, bytes)
		}
		llamp, ok = vibrationAmpLfaMap[llfAmp]
		if !ok {
			return fmt.Errorf("not found left low amplitude (%v): %+v", llfAmp, bytes)
		}
	}
	if !rightSkip {
		//var rhf uint16 = uint16(bytes[5]&0x01)<<8 | uint16(bytes[4])
		var rhfAmp uint8 = uint8(bytes[5] & 0xfe)
		//var rlf uint8 = uint8(bytes[6] & 0x7f)
		var rlfAmp uint16 = uint16(bytes[6]&0x80)<<8 | uint16(bytes[7])
		rhamp, ok = vibrationAmpHfaMap[rhfAmp]
		if !ok {
			return fmt.Errorf("not found right high amplitude (%v): %+v", rhfAmp, bytes)
		}
		rlamp, ok = vibrationAmpLfaMap[rlfAmp]
		if !ok {
			return fmt.Errorf("not found right low amplitude (%v): %+v", rlfAmp, bytes)
		}
	}
	if lhamp == 0 && llamp == 0 && rhamp == 0 && rlamp == 0 {
		return nil
	}
	if n.verbose {
		log.Printf("lhamp = %v, llamp = %v, rhamp = %v, rlamp = %v", lhamp, llamp, rhamp, rlamp)
	}
	hamp := float64(lhamp + rhamp) / 2.0 / 1000.0
	lamp := float64(llamp + rlamp) / 2.0 / 1000.0
	if hamp > 1 {
		hamp = 1.0
	}
	if lamp > 1 {
		lamp = 1.0
	}
	vibrationMessage := &message.GamepadVibration {
		Duration:        1000,
		StartDelay:      0,
		StrongMagnitude: hamp,
		WeakMagnitude:   lamp,
	}
	n.SendVibration(vibrationMessage)
	return nil
}

func (n *NSProCon) buildAck(subCmd byte, existsReportData bool) byte {
	ack := byte(0x80)
	if existsReportData {
		ack |= subCmd
	}
	return ack
}

func (n *NSProCon) readReportLoop(f * os.File) {
	// usb reset magic 
	n.writeReport(f, usbReportIdOutput81, []byte{ 0x01, 0x00, 0x03 })
	buf := make([]byte, 64)
	for {
		select {
		case <-n.stopCh:
			return
		default:
		}
		rl, err := f.Read(buf)
		if err != nil {
			log.Printf("can not read request report from gadget device file: %v", err)
			return
		}
		if n.verbose {
			log.Printf("read %x", buf[:rl])
		}
		switch buf[0] {
		case usbReportIdInput80:
			switch buf[1] {
			case subTypeRequestMac:
				reportBytes := []byte{ buf[1], 0x00 /* padding */, usbDeviceTypeProController }
				reportBytes = append(reportBytes, n.macAddr...)
				err = n.writeReport(f, usbReportIdOutput81, reportBytes)
				if err != nil {
					log.Printf("can not write reponse report (81) to gadget device file: %v", err)
					return
				}
				n.comState = comStateMac
			case subTypeHandshake:
				err = n.writeReport(f, usbReportIdOutput81, []byte{ buf[1] })
				if err != nil {
					log.Printf("can not write reponse report (81) to gadget device file: %v", err)
					return
				}
				if n.comState == comStateBaudRate  {
					n.comState = comStateHandshake2
				} else {
					n.comState = comStateHandshake
				}
			case subTypeBaudRate:
				err = n.writeReport(f, usbReportIdOutput81, []byte{ buf[1] })
				if err != nil {
					log.Printf("can not write reponse report (81) to gadget device file: %v", err)
					return
				}
				n.comState = comStateBaudRate
			case subTypeDisableUsbTimeout:
				n.usbTimeout = false
				n.comState = comStateDisableUsbTimeout
				log.Printf("diable usb timeout")
			case subTypeEnableUsbTimeout:
				n.usbTimeout = true
				n.comState = comStateEnableUsbTimeout
			default:
				log.Printf("unsupported sub type (%x): %x", buf[1], buf[2:rl])
			}
		case reportIdInput01:
			// XXX buf[1]  = counter : What should i do?
			err = n.sendVibrationRequest(buf[2:10])
			if err != nil {
				log.Printf("can not forward vibration report (01) to user: %v", err)
			}
			switch buf[10] {
			case subCommandBluetoothManualPairing:
				// buf[11:43] ???
				// skip 0x81 01 01, 0x81 01 02
				// last response only
				ack := n.buildAck(subCommandBluetoothManualPairing, true)
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), ack, subCommandBluetoothManualPairing,
					[]byte{ 0x03 } ))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", ack, subCommandBluetoothManualPairing, err)
					return
				}
			case subCommandRequestDeviceInfo:
				ack := n.buildAck(subCommandRequestDeviceInfo, true)
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), ack, subCommandRequestDeviceInfo,
					[]byte{ 0x03, 0x48, 0x03, 0x02 }, n.reverseMacAddr, []byte{ 0x03 /* ??? */, 0x02 /* default */ } ))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", ack, subCommandRequestDeviceInfo, err)
					return
				}
			case subCommandSetInputReportMode:
				if buf[11] == 0x30 {
					if n.verbose {
						log.Printf("Standard full mode. Pushes current state @60Hz")
					}
				}
				ack := n.buildAck(subCommandSetInputReportMode, false)
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), ack, subCommandSetInputReportMode))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", ack, subCommandSetInputReportMode, err)
					return
				}
			case subCommandTriggerButtonsElapsedTime:
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), 0x83 /* from dump */, subCommandTriggerButtonsElapsedTime))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", 0x83 /* from dump */, subCommandTriggerButtonsElapsedTime, err)
					return
				}
			case subCommandSetHciState:
				// buf[11] = 0x00 Disconnect 
				ack := n.buildAck(subCommandSetHciState, false)
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), ack, subCommandSetHciState))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", ack, subCommandSetHciState, err)
					return
				}
				// usb disconnect magic
				// n.writeReport(f, usbReportIdOutput81, []byte{ 0x01, 0x03 })
			case subCommandSetShipmentLowPowerState:
				// buf[11] nothig to do
				ack := n.buildAck(subCommandSetShipmentLowPowerState, false)
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), ack, subCommandSetShipmentLowPowerState))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", ack, subCommandSetShipmentLowPowerState, err)
					return
				}
			case subCommandReadSpi:
				var mem []byte
				switch buf[12] {
				case 0x60:
					mem = n.spiMemory60
				case 0x80:
					mem = n.spiMemory80
				default:
					log.Printf("unsupported spi memory address (%x:%x%x) to gadget device file: %v", subCommandReadSpi, buf[12], buf[11], err)
					err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), 0x00, subCommandReadSpi))
					if err != nil {
						log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", 0x00, subCommandReadSpi, err)
						return
					}
					return
				}
				if len(mem) < int(buf[11] + buf[15]) {
					log.Printf("unsupported spi memory address (%x:%x%x) to gadget device file: %v", subCommandReadSpi, buf[12], buf[11], err)
					err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), 0x00, subCommandReadSpi))
					if err != nil {
						log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", 0x00, subCommandReadSpi, err)
						return
					}
				}
				ack := n.buildAck(subCommandReadSpi, true)
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(),
					 ack, subCommandReadSpi, buf[11:16], mem[buf[11]:buf[11] + buf[15]]))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", ack, subCommandReadSpi, err)
					return
				}
                        case subCommandSetNfcIrMcuConfiguration:
				// ignore????
				// XXX buf[11] ????
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), 0xa0 /* from dump */, subCommandSetNfcIrMcuConfiguration,
				       []byte{ 0x01, 0x00, 0xff, 0x00, 0x03, 0x00, 0x05, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					       0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5c, /* XXX ??? */} ))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", 0xa0 /* from dump */, subCommandSetNfcIrMcuConfiguration, err)
					return
				}
			case subCommandSetPlayerLights:
				// buf[11] nothig to do
				ack := n.buildAck(subCommandSetPlayerLights, false)
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), ack, subCommandSetPlayerLights))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", ack, subCommandSetPlayerLights, err)
					return
				}
			case subCommand33:
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), 0x80 /* from dump */, subCommand33, []byte{ 0x03 /* XXX ???? */ }))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", 0x80 /* from dump */, subCommand33, err)
					return
				}
			case subCommandSetHomeLight:
				// buf[11:36] nothing to do
				ack := n.buildAck(subCommandSetHomeLight, false)
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), ack, subCommandSetHomeLight))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", ack, subCommandSetHomeLight, err)
					return
				}
			case subCommandEnableImu:
				n.imuEnable = buf[11]
				ack := n.buildAck(subCommandEnableImu, false)
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), ack, subCommandEnableImu))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", ack, subCommandEnableImu, err)
					return
				}
			case subCommandSetImuSensitivity:
                                n.imuSensitivity.gyroSensitivity              = buf[11]
                                n.imuSensitivity.accelerometerSensitivity     = buf[12]
                                n.imuSensitivity.gyroPerformanceRate          = buf[13]
                                n.imuSensitivity.accelerometerFilterBandwidth = buf[14]
				ack := n.buildAck(subCommandSetImuSensitivity, false)
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), ack, subCommandSetImuSensitivity))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", ack, subCommandSetImuSensitivity, err)
					return
				}
			case subCommandEnableVibration:
				n.vibrationEnable = buf[11]
				if n.verbose {
					if n.vibrationEnable == 0 {
						log.Printf("vibration disabled")
					} else {
						log.Printf("vibration enabled")
					}
				}
				ack := n.buildAck(subCommandEnableVibration, false)
				err = n.writeReport(f, reportIdOutput21, n.buildOutput21(n.buildControllerReport(), ack, subCommandEnableVibration))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %v", ack, subCommandEnableVibration, err)
					return
				}
			default:
				log.Printf("unsupported sub command (%x): %x", buf[10], buf[11:rl])
			}
		case reportIdInput10:
			// XXX buf[1]  = counter : What should i do?
			err = n.sendVibrationRequest(buf[2:10])
			if err != nil {
				log.Printf("can not forward vibration report (10) to user: %v", err)
			}
		}
	}
}

func (n *NSProCon) buildControllerReport() []byte {
        now := time.Now()
        timestamp := byte(((now.UnixNano() / int64(time.Millisecond)) % 256))
	byte1 := byte(8) /* buttery full */            << 4 |
	         byte(1) /* XXX connection info ??? */
	byte2 := n.controller.buttons.y            |
		 n.controller.buttons.x       << 1 |
		 n.controller.buttons.b       << 2 |
		 n.controller.buttons.a       << 3 |
		 n.controller.buttons.rightSr << 4 |
		 n.controller.buttons.rightSl << 5 |
		 n.controller.buttons.r       << 6 |
		 n.controller.buttons.zr      << 7
	byte3 := n.controller.buttons.minus             |
	         n.controller.buttons.plus         << 1 |
	         n.controller.rightStick.press     << 2 |
	         n.controller.leftStick.press      << 3 |
	         n.controller.buttons.home         << 4 |
	         n.controller.buttons.capture      << 5 |
	         0 /* unused */                    << 6 |
	         n.controller.buttons.chargingGrip << 7
	byte4 := n.controller.buttons.down        |
		 n.controller.buttons.up     << 1 |
		 n.controller.buttons.right  << 2 |
		 n.controller.buttons.left   << 3 |
		 n.controller.buttons.leftSr << 4 |
		 n.controller.buttons.leftSl << 5 |
		 n.controller.buttons.l      << 6 |
		 n.controller.buttons.zl     << 7
	lx := uint16(math.Round((1 + n.controller.leftStick.x) * 2047.5))
	ly := uint16(math.Round((1 + n.controller.leftStick.y) * 2047.5))
	rx := uint16(math.Round((1 + n.controller.rightStick.x) * 2047.5))
	ry := uint16(math.Round((1 + n.controller.rightStick.y) * 2047.5))
	// 0 - 4095 (12 bit)
	// 16 bit 8byte -> 12bit 6byte
	stickBytes := make([]byte, 6)
	stickBytes[0] = uint8(lx & 0xff)
	stickBytes[1] = uint8(((ly << 4) & 0xf0) | ((lx >> 8) & 0x0f))
	stickBytes[2] = uint8((ly >> 4) & 0xff)
	stickBytes[3] = uint8(rx & 0xff)
	stickBytes[4] = uint8(((ry << 4) & 0xf0) | ((rx >> 8) & 0x0f))
	stickBytes[5] = uint8((ry >> 4) & 0xff)
	vibratorReport := uint8(0x00) /* XXX ???? */
	return []byte{
		timestamp, byte1, byte2, byte3, byte4,
	        stickBytes[0], stickBytes[1], stickBytes[2],
	        stickBytes[3], stickBytes[4], stickBytes[5],
		vibratorReport,
	 }
}

func (n *NSProCon) buildOutput21(controller []byte, ack byte, subCmd byte, dataList ...[]byte) []byte {
	report := append(controller, ack, subCmd)
	for _, data := range dataList {
		report = append(report, data...)
	}
	return report
}

func (n *NSProCon) buildOutput30() []byte {
	report := n.buildControllerReport()
        if n.imuEnable != 0 {
                // XXX  not supported imu in gamepad api
                // XXX  no idea
		// XXX  report = append(report, imu...)
        }
	return report
}

func (n *NSProCon) writeControllerReportLoop(f *os.File) {
	ticker := time.NewTicker(time.Millisecond * 1000 / 60)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if n.comState < comStateDisableUsbTimeout {
				continue
			}
			err := n.writeReport(f, reportIdOutput30, n.buildControllerReport())
			if err != nil {
				log.Printf("can not write report (30) to gadget device file: %v", err)
				return
			}
		case <-n.stopCh:
			return
		}
	}
}

func (n *NSProCon) Setup() error {
	err := setup.UsbGadgetHidCleanup(n.setupParams)
	if err != nil {
		return fmt.Errorf("can not cleanup usb gadget hid device in nsprocon: %w", err)
	}
	time.Sleep(time.Second)
	err = setup.UsbGadgetHidSetup(n.setupParams)
	if err != nil {
		return fmt.Errorf("can not setup usb gadget hid device in nsprocon: %w", err)
	}
	err = setup.UsbGadgetHidEnable(n.setupParams)
	if err != nil {
		return fmt.Errorf("can not enable usb gadget hid device in nsprocon: %w", err)
	}
	time.Sleep(time.Second)
	_, err =  os.Stat("/dev/hidg0")
	if err != nil {
		return fmt.Errorf("not found device file  (%v) in nsprocon: %w", n.devFilePath, err)
	}
	return nil
}

func (n *NSProCon) Start() error {
        f, err := os.OpenFile(n.devFilePath, os.O_RDWR, 0644)
        if err != nil {
                return fmt.Errorf("can not open device file (%v) in nsprocon: %w", n.devFilePath, err)
        }
	n.devFile = f
	go n.readReportLoop(f)
	go n.writeControllerReportLoop(f)
	return nil
}

func (n *NSProCon) Stop() {
	close(n.stopCh)
	n.devFile.Close()
	err := setup.UsbGadgetHidDisable(n.setupParams)
	if err != nil {
		log.Printf("can not disable usb gadget hid device in nsprocon: %v", err)
	}
	err = setup.UsbGadgetHidCleanup(n.setupParams)
	if err != nil {
		log.Printf("can not cleanup usb gadget hid device in nsprocon: %v", err)
	}
}

func (n *NSProCon) boolToByte(v bool) byte {
	if v {
		return byte(1)
	} else {
		return byte(0)
	}
}

func (n *NSProCon) UpdateState(state *message.GamepadState) error {
	for i, button := range state.Buttons {
		switch i {
		case 0:
			n.controller.buttons.b = n.boolToByte(button.Pressed)
		case 1:
			n.controller.buttons.a = n.boolToByte(button.Pressed)
		case 2:
			n.controller.buttons.y = n.boolToByte(button.Pressed)
		case 3:
			n.controller.buttons.x = n.boolToByte(button.Pressed)
		case 4:
			n.controller.buttons.l = n.boolToByte(button.Pressed)
		case 5:
			n.controller.buttons.r = n.boolToByte(button.Pressed)
		case 6:
			n.controller.buttons.zl = n.boolToByte(button.Pressed)
		case 7:
			n.controller.buttons.zr = n.boolToByte(button.Pressed)
		case 8:
			n.controller.buttons.minus = n.boolToByte(button.Pressed)
		case 9:
			n.controller.buttons.plus = n.boolToByte(button.Pressed)
		case 10:
			n.controller.leftStick.press = n.boolToByte(button.Pressed)
		case 11:
			n.controller.rightStick.press = n.boolToByte(button.Pressed)
		case 12:
			n.controller.buttons.up = n.boolToByte(button.Pressed)
		case 13:
			n.controller.buttons.down = n.boolToByte(button.Pressed)
		case 14:
			n.controller.buttons.left = n.boolToByte(button.Pressed)
		case 15:
			n.controller.buttons.right = n.boolToByte(button.Pressed)
		case 16:
			n.controller.buttons.home = n.boolToByte(button.Pressed)
		case 17:
			n.controller.buttons.capture = n.boolToByte(button.Pressed)
		default:
			log.Printf("can not update state because unsupported button in nsprocon: button index = %v", i)
		}
	}
	for i, axis := range state.Axes {
		switch i {
		case 0:
			n.controller.leftStick.x = axis
		case 1:
			n.controller.leftStick.y = axis * -1.0
		case 2:
			n.controller.rightStick.x = axis
		case 3:
			n.controller.rightStick.y = axis * -1.0
		default:
			log.Printf("can not update state because unsupported axis in nsprocon: axis index = %v", i)
		}
	}
	if n.verbose {
		log.Printf("buttons = %+v, left stick = %+v, right stick = %+v",
			n.controller.buttons, n.controller.leftStick, n.controller.rightStick)
	}
	return nil
}

func (n *NSProCon) Press(buttons []ButtonName) error {
	for _, button := range buttons {
		switch button {
		case ButtonA:
			n.controller.buttons.a = 1
		case ButtonB:
			n.controller.buttons.b = 1
		case ButtonX:
			n.controller.buttons.x = 1
		case ButtonY:
			n.controller.buttons.y = 1
		case ButtonLeft:
			n.controller.buttons.left = 1
		case ButtonRight:
			n.controller.buttons.right = 1
		case ButtonUp:
			n.controller.buttons.up = 1
		case ButtonDown:
			n.controller.buttons.down = 1
		case ButtonPlus:
			n.controller.buttons.plus = 1
		case ButtonMinus:
			n.controller.buttons.minus = 1
		case ButtonHome:
			n.controller.buttons.home = 1
		case ButtonCapture:
			n.controller.buttons.capture = 1
		case ButtonStickL:
			n.controller.leftStick.press = 1
		case ButtonStickR:
			n.controller.rightStick.press = 1
		case ButtonL:
			n.controller.buttons.l = 1
		case ButtonR:
			n.controller.buttons.r = 1
		case ButtonZL:
			n.controller.buttons.zl = 1
		case ButtonZR:
			n.controller.buttons.zr = 1
		case ButtonLeftSL:
			n.controller.buttons.leftSl = 1
		case ButtonLeftSR:
			n.controller.buttons.leftSr = 1
		case ButtonRightSL:
			n.controller.buttons.rightSl = 1
		case ButtonRightSR:
			n.controller.buttons.rightSr = 1
		case ButtonChargingGrip:
			n.controller.buttons.chargingGrip = 1
		default:
			return fmt.Errorf("can not press because unsupported button in nsprocon: %v", button)
		}
	}
	return nil
}

func (n *NSProCon) Release(buttons []ButtonName) error {
	for _, button := range buttons {
		switch button {
		case ButtonA:
			n.controller.buttons.a = 0
		case ButtonB:
			n.controller.buttons.b = 0
		case ButtonX:
			n.controller.buttons.x = 0
		case ButtonY:
			n.controller.buttons.y = 0
		case ButtonLeft:
			n.controller.buttons.left = 0
		case ButtonRight:
			n.controller.buttons.right = 0
		case ButtonUp:
			n.controller.buttons.up = 0
		case ButtonDown:
			n.controller.buttons.down = 0
		case ButtonPlus:
			n.controller.buttons.plus = 0
		case ButtonMinus:
			n.controller.buttons.minus = 0
		case ButtonHome:
			n.controller.buttons.home = 0
		case ButtonCapture:
			n.controller.buttons.capture = 0
		case ButtonStickL:
			n.controller.leftStick.press = 0
		case ButtonStickR:
			n.controller.rightStick.press = 0
		case ButtonL:
			n.controller.buttons.l = 0
		case ButtonR:
			n.controller.buttons.r = 0
		case ButtonZL:
			n.controller.buttons.zl = 0
		case ButtonZR:
			n.controller.buttons.zr = 0
		case ButtonLeftSL:
			n.controller.buttons.leftSl = 0
		case ButtonLeftSR:
			n.controller.buttons.leftSr = 0
		case ButtonRightSL:
			n.controller.buttons.rightSl = 0
		case ButtonRightSR:
			n.controller.buttons.rightSr = 0
		case ButtonChargingGrip:
			n.controller.buttons.chargingGrip = 0
		default:
			return fmt.Errorf("can not release because unsupported button in nsprocon: %v", button)
		}
	}
	return nil
}

func (n *NSProCon) StickL(xAxis float64, yAxis float64) error {
	n.controller.leftStick.x = xAxis
	n.controller.leftStick.y = yAxis * -1.0
	return nil
}

func (n *NSProCon) StickR(xAxis float64, yAxis float64) error {
	n.controller.rightStick.x = xAxis
	n.controller.rightStick.y = yAxis * -1.0
	return nil
}

func NewNSProCon(verbose bool, macAddr string, spiMemory60 string, spiMemory80 string, devFilePath string, configsHome string, udc string) (*NSProCon, error) {
	setupParams := &setup.UsbGadgetHidSetupParams{
		ConfigsHome:     configsHome,
		GadgetName:      "nsprocon",
		IdProduct:       "0x2009",
		IdVendor:        "0x057e",
		BcdDevice:       "0x0200",
		BcdUsb:          "0x0200",
		BMaxPacketSize0: "64",
		BDeviceProtocol: "0",
		BDeviceSubClass: "0",
		BDeviceClass:    "0",
		StringsLang:     "0x409",
		ISerial:         "000000000001",
		IProduct:        "Pro Controller",
		IManufacturer:   "Nintendo Co., Ltd.",
		ConfigName:      "c",
		ConfigNumber:    "1",
		ConfigString:    "Nintendo Switch Pro Controller",
		BmAttributes:    "0xa0",
		MaxPower:        "500",
		FunctionName:    "hid",
		InstanceName:    "usb0",
		Protocol:        "0",
		Subclass:        "0",
		ReportLength:    "203",
		ReportDesc:      "050115000904A1018530050105091901290A150025017501950A5500650081020509190B290E150025017501950481027501950281030B01000100A1000B300001000B310001000B320001000B35000100150027FFFF0000751095048102C00B39000100150025073500463B0165147504950181020509190F2912150025017501950481027508953481030600FF852109017508953F8103858109027508953F8103850109037508953F9183851009047508953F9183858009057508953F9183858209067508953F9183C0",
		UDC:	         udc,
	}
	defaultDevFilePath := "/dev/hidg0"
	if devFilePath == "" {
		devFilePath = defaultDevFilePath
	}
        decodedMacAddr, err := hex.DecodeString(macAddr)
        if err != nil {
                return nil, fmt.Errorf("can not decode mac address string (%v): %w", macAddr, err)
        }
        decodedSpiMemory60, err := hex.DecodeString(spiMemory60)
        if err != nil {
                return nil, fmt.Errorf("can not decode spi memory 60XX string (%v): %w", decodedSpiMemory60, err)
        }
        decodedSpiMemory80, err := hex.DecodeString(spiMemory80)
        if err != nil {
                return nil, fmt.Errorf("can not decode spi memory 80XX string (%v): %w", decodedSpiMemory80, err)
        }
	reverseMacAddr := make([]byte, len(decodedMacAddr))
	for i, b := range decodedMacAddr {
		reverseMacAddr[len(decodedMacAddr) - 1 - i] = b
	}
	return &NSProCon{
		BaseBackend: &BaseBackend{
			verbose: verbose,
		},
		verbose: verbose,
		setupParams: setupParams,
		macAddr: decodedMacAddr,
		reverseMacAddr: reverseMacAddr,
		spiMemory60: decodedSpiMemory60,
		spiMemory80: decodedSpiMemory80,
		devFilePath: devFilePath,
		devFile: nil,
		comState: comStateInit,
		usbTimeout: true,
		reportCounter: 0,
		imuEnable: 0,
		vibrationEnable: 0,
		stopCh: make(chan int),
		controller: &controller{
			buttons: &buttons{},
			leftStick: &stick{},
			rightStick: &stick{},
		},
		imuSensitivity: &imuSensitivity{},
	}, nil
}
