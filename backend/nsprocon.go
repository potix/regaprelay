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
	comStateInit       comState = iota
	comStateMac
	comStateBaudRate
	comStateHandshake
)


// Report IDs.
const byte reportIdInput01 = 0x01;
const byte reportIdInput10 = 0x10;
const byte reportIdOutput21 = 0x21;
const byte reportIdOutput30 = 0x30;
const byte usbReportIdInput80 = 0x80;
const byte usbReportIdOutput81 = 0x81;

// Sub-types of the 0x81 input report, used for initialization.
const byte subTypeRequestMac = 0x01;
const byte subTypeHandshake = 0x02;
const byte subTypeBaudRate = 0x03;
const byte subTypeDisableUsbTimeout = 0x04;
const byte subTypeEnableUsbTimeout = 0x05;

// Values for the |device_type| field reported in the MAC reply.
const byte usbDeviceTypeChargingGripNoDevice = 0x00;
const byte usbDeviceTypeChargingGripJoyConL = 0x01;
const byte usbDeviceTypeChargingGripJoyConR = 0x02;
const byte subDeviceTypeProController = 0x03;

type NSProCon struct  {
	*BaseBackend
	setupParams    *setup.UsbGadgetHidSetupParams
	verbose        bool
	macAddr        []byte
	reverseMacAddr []byte
	comState       comState
	usbTimeout     bool
	reportCounter  byte
	imuEnable      bool
	stopCh         int
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
	return nil
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
				err := writeReport(kUsbReportIdOutput81, reportBytes)
				if err != nil {
					log.Printf("can not write reponse report (81) to gadget device file: %w", err)
					return
				}
				n.comState = comStateMac
			case subTypeHandshake:
				err := writeReport(kUsbReportIdOutput81, []byte{ buf[1] })
				if err != nil {
					log.Printf("can not write reponse report (81) to gadget device file: %w", err)
					return
				}
				n.comState = comStateHandshake
			case subTypeBaudRate:
				err := writeReport(kUsbReportIdOutput81, []byte{ buf[1] })
				if err != nil {
					log.Printf("can not write reponse report (81) to gadget device file: %w", err)
					return
				}
				n.comState = comStateBaudRate
			case subTypeDisableUsbTimeout:
				n.usbTimeout = false
			case subTypeEnableUsbTimeout:
				n.usbTimeout = true
			default:
				log.Printf("unsupported sub type (%x): %x", buf[1], buf[2:rl])
			}
		case kUsbReportIdInput01:
		case kUsbReportIdInput10:
		}
	}
}

func (n *NSProCon) buildControllerStateReport() []byte {
        now := time.Now()
        timestamp := ((now.UnixNano() / int64(time.Millisecond)) % 256)
	byte1 := 8 /* buttery full */               << 4 |
	         1 /* connection info ??? */
	byte2 := c.controllerState.buttons.y            |
		 c.controllerState.buttons.x       << 1 |
		 c.controllerState.buttons.b       << 2 |
		 c.controllerState.buttons.a       << 3 |
		 c.controllerState.buttons.rightSr << 4 |
		 c.controllerState.buttons.rightSl << 5 |
		 c.controllerState.buttons.r       << 6 |
		 c.controllerState.buttons.zr      << 7
	byte3 := c.controllerState.buttons.minus             |
	         c.controllerState.buttons.plus         << 1 |
	         c.controllerState.rightStick.press     << 2 |
	         c.controllerState.leftStick.press      << 3 |
	         c.controllerState.buttons.home         << 4 |
	         c.controllerState.buttons.capture      << 5 |
	         0 /* unused */                         << 6 |
	         c.controllerState.buttons.chargingGrip << 7 |
	byte4 := c.controllerState.buttons.down        |
		 c.controllerState.buttons.up     << 1 |
		 c.controllerState.buttons.right  << 2 |
		 c.controllerState.buttons.left   << 3 |
		 c.controllerState.buttons.leftSr << 4 |
		 c.controllerState.buttons.leftSl << 5 |
		 c.controllerState.buttons.l      << 6 |
		 c.controllerState.buttons.zl     << 7 |
	lx := uint16(math.Round((1 + c.controllerState.leftStick.x) * 2047.5))
	ly := uint16(math.Round((1 + c.controllerState.leftStick.y) * 2047.5))
	rx := uint16(math.Round((1 + c.controllerState.rightStick.x) * 2047.5))
	ry := uint16(math.Round((1 + c.controllerState.rightStick.y) * 2047.5))
	// 0 - 4095 (12 bit)
	// 16 bit 8byte -> 12bit 6byte
	stickBytes = make([]byte, 6)
	stickBytes[0] = uint8(lx & 0xff)
	stickBytes[1] = uint8(((ly << 4) & 0xf0) | ((lx >> 8) & 0x0f))
	stickBytes[2] = uint8((ly >> 4) & 0xff)
	stickBytes[3] = uint8(rx & 0xff)
	stickBytes[4] = uint8(((ry << 4) & 0xf0) | ((rx >> 8) & 0x0f))
	stickBytes[5] = uint8((ry >> 4) & 0xff)
	vibratorReport := uint8(0x00) /* ???? */
	if c.imuEnable {
		// XXX  not supported imu in gamepad api
		// XXX  no idea
	}
	return []byte{timestamp, byte1, byte2, byte3, byte4,
		stickBytes[0], stickBytes[1], stickBytes[2],
	        stickBytes[3], stickBytes[4], stickBytes[5],
		vibratorReport }
}

func (n *NSProCon) writeControllerStateReportLoop() {
	ticker := time.NewTicker(time.Millisecond * 1000 / 60)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if n.comState < comStateHandshake {
				continue
			}
			err := c.writeReport(reportIdOutput30, buildControllerStateReport() )
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
		imuEnable: false,
		stopCha: make(chan int),
	}, nil
}
