package backend

import(
	"fmt"
	"log"
	"time"
	"github.com/potix/regaprelay/backend/setup"
	"github.com/potix/regapweb/handler"
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
const byte reportIdInput01     = 0x01
const byte reportIdInput10     = 0x10
const byte reportIdOutput21    = 0x21
const byte reportIdOutput30    = 0x30
const byte usbReportIdInput80  = 0x80
const byte usbReportIdOutput81 = 0x81

// Sub-types of the 0x81 input report, used for initialization.
const byte subTypeRequestMac        = 0x01
const byte subTypeHandshake         = 0x02
const byte subTypeBaudRate          = 0x03
const byte subTypeDisableUsbTimeout = 0x04
const byte subTypeEnableUsbTimeout  = 0x05

// Values for the |device_type| field reported in the MAC reply.
const byte usbDeviceTypeChargingGripNoDevice = 0x00
const byte usbDeviceTypeChargingGripJoyConL  = 0x01
const byte usbDeviceTypeChargingGripJoyConR  = 0x02
const byte usbDeviceTypeProController        = 0x03

// UART subcommands.
const byte subCommandBluetoothManualPairing   = 0x01
const byte subCommandRequestDeviceInfo        = 0x02
const byte subCommandSetInputReportMode       = 0x03
const byte subCommandSetShipmentLowPowerState = 0x08
const byte subCommandReadSpi                  = 0x10
const byte subCommandSetNfcIrMcuConfiguration = 0x21
const byte subCommandSetPlayerLights          = 0x30
const byte subCommand33                       = 0x33
const byte subCommandSetHomeLight             = 0x38
const byte subCommandEnableImu                = 0x40
const byte subCommandSetImuSensitivity        = 0x41
const byte subCommandEnableVibration          = 0x48

vibrationAmpHfaMap := map[uint8]int{
    0x00: 0,   0x02: 10,   0x04: 12,
    0x06: 14,  0x08: 17,   0x0a: 20,
    0x0c: 24,  0x0e: 28,   0x10: 33,
    0x12, 40,  0x14: 47,   0x16: 56,
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

vibrationAmpLfaMap := map[uint16]int{
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
	comState        comState
	usbTimeout      bool
	reportCounter   byte
	imuEnable       byte
	vibrationEnable byte
	stopCh          int
	controller      *controller
	imuSensitivity  *imuSensitivity
}

func (n *NSProCon) writeReport(f *File, reportId byte, reportBytes []byte) (error) {
	buf := make([]byte, 64)
	buf[0] = reportId
	for i, b := range reportByte {
		buf[i + 1] = b
	}
	wl, err := f.Write(buf)
	if err != nil {
		return fmt.Errorf("can not write report (%x) to gadget device file: %w", reportId,  err)
	}
	if wl != len(buf) {
		return fmt.Errorf("partial write report (%x) to gadget device file: write len = %v", reportId, wl)
	}
	log.Printf("wrote %x", buf)
	return nil
}

func (n **NSProCon) sendVibrationRequest(bytes []byte) error {
	if len(bytes) < 8 {
		return fmt.Errorf("invalid vibration data (%x)", bytes)
	}
	if bytes[0] == 0 && bytes[1] == 0 && bytes[2] == 0 && bytes[3] == 0 &&
	   bytes[4] == 0 && bytes[5] == 0 && bytes[6] == 0 && bytes[7] == 0 {
		return nil
	}
	var lhf uint16 = uint16(bytes[1]&0x01)<<8 | uint16(bytes[0])
	var lhfAmp uint8 = uint8(bytes[1] & 0xfe)
	var llf uint8 = uint8(bytes[2] & 0x7f)
	var llfAmp uint16 = uint16(bytes[2]&0x80)<<8 | uint16(bytes[3])
	var rhf uint16 = uint16(bytes[5]&0x01)<<8 | uint16(bytes[4])
	var rhfAmp uint8 = uint8(bytes[5] & 0xfe)
	var rlf uint8 = uint8(bytes[6] & 0x7f)
	var rlfAmp uint16 = uint16(bytes[6]&0x80)<<8 | uint16(bytes[7])
	log.Printf("lhf = %v, lhfAmp = %v, llf = %v, llfAmp = %v", lhf, lhfAmp, llf, llfAmp)
	log.Printf("rhf = %v, rhfAmp = %v, rlf = %v, rlfAmp = %v", rhf, rhfAmp, rlf, rlfAmp)
	lhamp, ok := vibrationAmpHfaMap[lhfAmp]
	if !ok {
		return fmt.Errorf("mot found amplitude (%v)", lhfAmp)
	}
	llamp, ok := vibrationAmpLfaMap[llfAmp]
	if !ok {
		return fmt.Errorf("mot found amplitude (%v)", llfAmp)
	}
	rhamp, ok := vibrationAmpHfaMap[lhfAmp]
	if !ok {
		return fmt.Errorf("mot found amplitude (%v)", lhfAmp)
	}
	rlamp, ok := vibrationAmpLfaMap[llfAmp]
	if !ok {
		return fmt.Errorf("mot found amplitude (%v)", llfAmp)
	}
	if lhamp == 0 && llamp == 0 && rhamp == 0 && rlamp == 0 {
		return nil
	}
	hamp := float64(lhamp + rhamp) / 2.0 / 1000.0
	lamp := float64(llamp + rlamp) / 2.0 / 1000.0
	vibrationMessage := &handler.GamepadVibrationMessage {
		Duration:        1000,
		StartDelay:      0,
		StrongMagnitude: hamp,
		WeakMagnitude:   lamp,
	}
	if n.onVibrationCh != nil {
		n.onVibrationCh <- vibrationMessage
	}
}

func (n *NSProCon) readReportLoop(f * File) {
	buf := make([]byte, 64)
	for {
		select {
		case <-c.stopCh:
			return
		default:
		}
		rl, err := f.Read(buf)
		if err != nil {
			log.Printf("can not read request report from gadget device file: %w", err)
			return
		}
		select buf[0] {
		case usbReportIdInput80:
			select buf[1] {
			case subTypeRequestMac:
				reportBytes := []byte{ buf[1], 0x00 /* padding */, subDeviceTypeProController }
				reportBytes = append(reportBytes, n.macAddr...)
				err = writeReport(usbReportIdOutput81, reportBytes)
				if err != nil {
					log.Printf("can not write reponse report (81) to gadget device file: %w", err)
					return
				}
				n.comState = comStateMac
			case subTypeHandshake:
				err = writeReport(usbReportIdOutput81, []byte{ buf[1] })
				if err != nil {
					log.Printf("can not write reponse report (81) to gadget device file: %w", err)
					return
				}
				if n.comState == comStateBaudRate  {
					n.comState = comStateHandshake2
				} else {
					n.comState = comStateHandshake
				}
			case subTypeBaudRate:
				err = writeReport(usbReportIdOutput81, []byte{ buf[1] })
				if err != nil {
					log.Printf("can not write reponse report (81) to gadget device file: %w", err)
					return
				}
				n.comState = comStateBaudRate
			case subTypeDisableUsbTimeout:
				n.usbTimeout = false
				n.comState = comStateDisableUsbTimeout
			case subTypeEnableUsbTimeout:
				n.usbTimeout = true
				n.comState = comStateEnableUsbTimeout
			default:
				log.Printf("unsupported sub type (%x): %x", buf[1], buf[2:rl])
			}
		case usbReportIdInput01:
			// XXX buf[1]  = counter : What should i do?
			err = n.sendVibrationRequest(buf[2:10])
			if err != nil {
				log.Printf("can not forward vibration report (01) to user: %w", err)
			}
			switch buf[10] {
			case subCommandBluetoothManualPairing:
				// XXX ignore
				log.Printf("ignore bluetooth manual pairing")
			case subCommandRequestDeviceInfo:
				err = writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x82 /* ack */, subCommandSetInputReportMode,
					[]byte{ 0x03, 0x48, 0x03, 0x02 /* ??? */ }, c.reverseMacAddr, []byte{ 0x03 /* ??? */, 0x02 /* default */ } ))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandSetInputReportMode, err)
					return
				}
			case subCommandSetInputReportMode:
				if buf[11] == x30 {
					log.Printf("Standard full mode. Pushes current state @60Hz")
				}
				err = writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandSetInputReportMode))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandSetInputReportMode, err)
					return
				}
			case subCommandSetShipmentLowPowerState:
				// buf[11] nothig to do
				err = writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandSetShipmentLowPowerState))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandSetShipmentLowPowerState, err)
					return
				}
			case subCommandReadSpi:
				var mem []byte
				switch buf[12]:
				case 0x60:
					mem = n.spiMemory60
				case 0x80:
					mem = n.spiMemory80
				default:
					log.Printf("unsupported spi memory (%x:%x%v) to gadget device file: %w", 0x80, subCommandReadSpi, buf[12], buf[11] err)
					return
				}
				err = writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(),
					 0x90 /* ack */, subCommandReadSpi, buf[11:16], mem[buf[11]:buf[11] + buf[15]]))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x90, subCommandReadSpi, err)
					return
				}
                        case subCommandSetNfcIrMcuConfiguration:
				// XXX buf[11] ????
				err = writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0xa0 /* ack */, subCommandSetNfcIrMcuConfiguration,
				       []byte{ 0x01, 0x00, 0xff, 0x00, 0x03, 0x00, 0x05, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					       0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5c, /* XXX ??? */} ))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0xa0, subCommandSetNfcIrMcuConfiguration, err)
					return
				}
			case subCommandSetPlayerLights:
				// buf[11] nothig to do
				err = writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandSetPlayerLights))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandSetPlayerLights, err)
					return
				}
			case subCommand33:
				err = writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommand33, []byte{ 0x03 /* XXX ???? */ }))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommand33, err)
					return
				}

			case subCommandSetHomeLight:
				// buf[11:36] nothing to do
				err = writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandSetHomeLight))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandSetHomeLight, err)
					return
				}
			case subCommandEnableImu:
				c.imuEnable := buf[11]
				err = writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandEnableImu))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandEnableImu, err)
					return
				}
			case subCommandSetImuSensitivity:
                                c.imuSensitivity.gyroSensitivity              = buf[11]
                                c.imuSensitivity.accelerometerSensitivity     = buf[12]
                                c.imuSensitivity.gyroPerformanceRate          = buf[13]
                                c.imuSensitivity.accelerometerFilterBandwidth = buf[14]
				err = writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandSetImuSensitivity))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandSetImuSensitivity, err)
					return
				}
			case subCommandEnableVibration:
				c.vibrationEnable = buf[11]
				if c.vibrationEnable == 0 {
					log.Printf("vibration disabled")
				} else {
					log.Printf("vibration enabled")
				}
				err = writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandEnableVibration))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandEnableVibration, err)
					return
				}
			default:
				log.Printf("unsupported sub command (%x): %x", buf[10], buf[11:rl])
			}
		case usbReportIdInput10:
			// XXX buf[1]  = counter : What should i do?
			err = n.sendVibrationRequest(buf[2:10])
			if err != nil {
				log.Printf("can not forward vibration report (10) to user: %w", err)
			}
		}
	}
}

func (n *NSProCon) buildControllerReport() []byte {
        now := time.Now()
        timestamp := ((now.UnixNano() / int64(time.Millisecond)) % 256)
	byte1 := 8 /* buttery full */               << 4 |
	         1 /* connection info ??? */
	byte2 := c.controller.buttons.y            |
		 c.controller.buttons.x       << 1 |
		 c.controller.buttons.b       << 2 |
		 c.controller.buttons.a       << 3 |
		 c.controller.buttons.rightSr << 4 |
		 c.controller.buttons.rightSl << 5 |
		 c.controller.buttons.r       << 6 |
		 c.controller.buttons.zr      << 7
	byte3 := c.controller.buttons.minus             |
	         c.controller.buttons.plus         << 1 |
	         c.controller.rightStick.press     << 2 |
	         c.controller.leftStick.press      << 3 |
	         c.controller.buttons.home         << 4 |
	         c.controller.buttons.capture      << 5 |
	         0 /* unused */                    << 6 |
	         c.controller.buttons.chargingGrip << 7 |
	byte4 := c.controller.buttons.down        |
		 c.controller.buttons.up     << 1 |
		 c.controller.buttons.right  << 2 |
		 c.controller.buttons.left   << 3 |
		 c.controller.buttons.leftSr << 4 |
		 c.controller.buttons.leftSl << 5 |
		 c.controller.buttons.l      << 6 |
		 c.controller.buttons.zl     << 7 |
	lx := uint16(math.Round((1 + c.controller.leftStick.x) * 2047.5))
	ly := uint16(math.Round((1 + c.controller.leftStick.y) * 2047.5))
	rx := uint16(math.Round((1 + c.controller.rightStick.x) * 2047.5))
	ry := uint16(math.Round((1 + c.controller.rightStick.y) * 2047.5))
	// 0 - 4095 (12 bit)
	// 16 bit 8byte -> 12bit 6byte
	stickBytes = make([]byte, 6)
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
	report := append(controller, ack, subcmd)
	for _, data := range dataList {
		report = append(report, data...)
	}
	return report
}

func (n *NSProCon) buildOutput30() []byte {
	report := buildControllerReport()
        if c.imuEnable != 0 {
                // XXX  not supported imu in gamepad api
                // XXX  no idea
		// XXX  report = append(report, imu...)
        }
	return report
}

func (n *NSProCon) writeControllerReportLoop() {
	ticker := time.NewTicker((time.Millisecond * 1000 / 60) + time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if n.comState < comStateDisableUsbTimeout {
				continue
			}
			err := c.writeReport(reportIdOutput30, buildControllerReport())
		case <-c.stopCh:
			return
		}
	}
}

func (n *NSProCon) Setup() error {
	err := setup.UsbGadgetHidCleanup(n.setupParams)
	if err != nil {
		return fmt.Errorf("can not cleanup usb gadget hid device in nsprocon: %w", err)
	}
	time.Sleep(500 * time.Millisecond)
	err = setup.UsbGadgetHidSetup(n.setupParams)
	if err != nil {
		return fmt.Errorf("can not setup usb gadget hid device in nsprocon: %w", err)
	}
	err = setup.UsbGadgetHidEnable(n.setupParams)
	if err != nil {
		return fmt.Errorf("can not enable usb gadget hid device in nsprocon: %w", err)
	}
	time.Sleep(500 * time.Millisecond)
	return nil
}

func (n *NSProCon) Start() error {
	go n.readReportLoop()
	go n.writeControllerReportLoop()
	return nil
}

func (n *NSProCon) Stop() {
	close(n.stopCh)
	err := setup.UsbGadgetHidDisable(n.setupParams)
	if err != nil {
		log.Printf("can not disable usb gadget hid device in nsprocon: %v", err)
	}
	err = setup.UsbGadgetHidCleanup(n.setupParams)
	if err != nil {
		log.Printf("can not cleanup usb gadget hid device in nsprocon: %w", err)
	}
}

func (n *NSProCon) UpdateState(state *handler.GamepadStateMessage) error {
	// XXX TODO
	log.Printf("state = %+v", state)
	return nil
}

func (n *NSProCon) Press(buttons []ButtonName) error {
	// XXX TODO
	return nil
}

func (n *NSProCon) Release(buttons []ButtonName) error {
	// XXX TODO
	return nil
}

func (n *NSProCon) Push(buttons []ButtonName, duration time.Duration) error {
	// XXX TODO
	return nil
}

func (n *NSProCon) Repeat(buttons []ButtonName, interval time.Duration, duration time.Duration) error {
	// XXX TODO
	return nil
}

func (n *NSProCon) StickL(xDir XDirection, xPower float64, yDir YDirection, yPower float64, duration time.Duration) error {
	// XXX TODO
	return nil
}

func (n *NSProCon) StickR(xDir XDirection, xPower float64, yDir YDirection, yPower float64, duration time.Duration) error {
	// XXX TODO
	return nil
}

func (n *NSProCon) RotationStickL(xDir XDirection, lapTime time.Duration, power float64, duration time.Duration) error {
	// XXX TODO
	return nil
}

func (n *NSProCon) RotationStickR(xDir XDirection, lapTime time.Duration, power float64, duration time.Duration) error {
	// XXX TODO
	return nil
}

func NewNSProCon(verbose bool, macAddr string, spiMemory60 string, spiMemory80 string, configsHome string, udc string) (*NSProCon, error) {
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
		BaseBackend: &BaseBackend{},
		verbose: verbose,
		setupParams: setupParams,
		macAddr: decodedMacAddr,
		reverseMacAddr: reverseMacAddr,
		spiMemory60: decoededSpiMemory60,
		spiMemory80: decoededSpiMemory80,
		comState: comStateInit,
		usbtimeout: true,
		reportCounter: 0,
		imuEnable: 0,
		bybrationEnable: 0,
		stopCha: make(chan int),
		controller: &controller{},
		imuSensitivity: &imuSensitivity{},
	}, nil
}
