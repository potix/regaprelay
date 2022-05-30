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
const byte subDeviceTypeProController        = 0x03

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





var spiRomData = map[byte][]byte{
	0x60: []byte{
		/* 0x6000 */ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		/* 0x6010 */             0x03, 0xa0,                                           0x02            
		/* 0x6020 */ 0x05, 0x00, 0x53, 0x00, 0x94, 0x01, 0x00, 0x40, 0x00, 0x40, 0x00, 0x40, 0xf4, 0xff, 0x09, 0x00,
		/* 0x6030 */ 0x01, 0x00, 0xe7, 0x3b, 0xe7, 0x3b, 0xe7, 0x3b,                               0x05, 0x96, 0x63,
		/* 0x6040 */ 0xbc, 0xc7, 0x7a, 0x57, 0x46, 0x5e, 0x90, 0x87, 0x7b, 0x39, 0x86, 0x5e, 0xef, 0x15, 0x63, 0xff,
		/* 0x6050 */ 0x32, 0x32, 0x32, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00,
		/* 0x6060 */
		/* 0x6070 */
		/* 0x6080 */ 0x50, 0xfd, 0x00, 0x00, 0xc6, 0x0f, 0x0f, 0x30, 0x61, 0xae, 0x90, 0xd9, 0xd4, 0x14, 0x54, 0x41,
		/* 0x6090 */ 0x15, 0x54, 0xc7, 0x79, 0x9c, 0x33, 0x36, 0x63, 0x0f, 0x30, 0x61, 0xae, 0x90, 0xd9, 0xd4, 0x14,
		/* 0x60a0 */ 0x54, 0x41, 0x15, 0x54, 0xc7, 0x79, 0x9c, 0x33, 0x36, 0x63,

	},
	0x80: []byte{
		/* 0x8000 */
		/* 0x8010 */ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		/* 0x8020 */ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff
		/* 0x8030 */ 
	},
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

func (n **NSProCon) sendVibrationRequest(vibrationBytes []byte) error {
	if len(vibrationBytes) < 8 {
		return fmt.Errorf("invalid vibration data")
	}
	var sum byte
	for _, b := range vibrationBytes {
		sum += b
	}
	if sum == 0 {
		// no data
		return
	}
	var lhf uint16 = uint16(bytes[1]&0x01)<<8 | uint16(bytes[0])
	var lhfAmp uint8 = uint8(bytes[1] & 0xfe)
	var llf uint8 = uint8(bytes[2] & 0x7f)
	var llfAmp uint16 = uint16(bytes[2]&0x80)<<8 | uint16(bytes[3])
	var rhf uint16 = uint16(bytes[5]&0x01)<<8 | uint16(bytes[4])
	var rhfAmp uint8 = uint8(bytes[5] & 0xfe)
	var rlf uint8 = uint8(bytes[6] & 0x7f)
	var rlfAmp uint16 = uint16(bytes[6]&0x80)<<8 | uint16(bytes[7])
	// XXXXX  テーブルから振動強度取る
	// XXXXX  ノーマライズする
	// XXXXX  平均化していい感じにする
	// XXXXX  構造体に合わせる
	vibrationMessage := &handler.GamepadVibrationMessage {
		// XXXX
	}
	if n.onVibrationCh != nil {
		n.onVibrationCh <- vibrationMessage
	}
}

func (n *NSProCon) readReportLoop(f * File) {
	buf := make([]byte, 128)
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
				err := writeReport(usbReportIdOutput81, reportBytes)
				if err != nil {
					log.Printf("can not write reponse report (81) to gadget device file: %w", err)
					return
				}
				n.comState = comStateMac
			case subTypeHandshake:
				err := writeReport(usbReportIdOutput81, []byte{ buf[1] })
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
				err := writeReport(usbReportIdOutput81, []byte{ buf[1] })
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
			counter := buf[1] // XXX ???
			n.sendVibrationRequest(buf[2:10])
			switch buf[10] {
			case subCommandBluetoothManualPairing:
				// XXX ignore
				log.Printf("ignore bluetooth manual pairing")
			case subCommandRequestDeviceInfo:
				err := writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x82 /* ack */, subCommandSetInputReportMode,
					[]byte{ 0x03, 0x48, 0x03, 0x02 /* ??? */ }, c.reverseMacAddr, []byte{ 0x03 /* ??? */, 0x02 /* default */ } ))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandSetInputReportMode, err)
					return
				}
			case subCommandSetInputReportMode:
				if buf[11] == x30 {
					log.Printf("Standard full mode. Pushes current state @60Hz")
				}
				err := writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandSetInputReportMode))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandSetInputReportMode, err)
					return
				}
			case subCommandSetShipmentLowPowerState:
				// buf[11] nothig to do
				err := writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandSetShipmentLowPowerState))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandSetShipmentLowPowerState, err)
					return
				}
			case subCommandReadSpi:
				// XXXX TODO
				// XXXX
                        case subCommandSetNfcIrMcuConfiguration:
				// XXX buf[11] ????
				err := writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0xa0 /* ack */, subCommandSetNfcIrMcuConfiguration,
				       []byte{ 0x01, 0x00, 0xff, 0x00, 0x03, 0x00, 0x05, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					       0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5c, /* XXX ??? */} ))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0xa0, subCommandSetNfcIrMcuConfiguration, err)
					return
				}
			case subCommandSetPlayerLights:
				// buf[11] nothig to do
				err := writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandSetPlayerLights))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandSetPlayerLights, err)
					return
				}
			case subCommand33:
				err := writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommand33, []byte{ 0x03 /* XXX ???? */ }))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommand33, err)
					return
				}

			case subCommandSetHomeLight:
				// buf[11:36] nothing todo
				err := writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandSetHomeLight))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandSetHomeLight, err)
					return
				}
			case subCommandEnableImu:
				c.imuEnable := buf[11]
				err := writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandEnableImu))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandEnableImu, err)
					return
				}
			case subCommandSetImuSensitivity:
                                c.imuSensitivity.gyroSensitivity              = buf[11]
                                c.imuSensitivity.accelerometerSensitivity     = buf[12]
                                c.imuSensitivity.gyroPerformanceRate          = buf[13]
                                c.imuSensitivity.accelerometerFilterBandwidth = buf[14]
				err := writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandSetImuSensitivity))
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
				err := writeReport(usbReportIdOutput21, buildOutput21(buildControllerReport(), 0x80 /* ack */, subCommandEnableVibration))
				if err != nil {
					log.Printf("can not write reponse report (21:%x:%x) to gadget device file: %w", 0x80, subCommandEnableVibration, err)
					return
				}
			default:
				log.Printf("unsupported sub command (%x): %x", buf[10], buf[11:rl])
			}
		case usbReportIdInput10:
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
	         0 /* unused */                         << 6 |
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
	if c.imuEnable != 0 {
		// XXX  not supported imu in gamepad api
		// XXX  no idea
	}
	return []byte{
		timestamp, byte1, byte2, byte3, byte4,
	        stickBytes[0], stickBytes[1], stickBytes[2],
	        stickBytes[3], stickBytes[4], stickBytes[5],
		vibratorReport,
	 }
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

func NewNSProCon(verbose bool, macAddr string, configsHome string, udc string) (*NSProCon, error) {
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
