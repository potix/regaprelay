package gamepad

type NSProCon struct  {
	BaseBackend
	setupParams *setup.UsbGadgetHidSetupParams
	verbose     bool
}

func (n *NSProCon) Setup() error {
	err := UsbGadgetHidSetup(n.setupParams)
	if err != nill {
		return fmt.Errorf("setup error in nsprocon: %w", err)
	}
	return nil
}

func (n *NSProCon) Start() error {
	err := UsbGadgetHidEnable(n.setupParams)
	if err != nill {
		return fmt.Errorf("can not enable usb gadget hid device in nsprocon: %w", err)
	}
}

func (n *NSProCon) Stop() {
	err := UsbGadgetHidDisable(n.setupParams)
	if err != nill {
		log.Printf("can not enable usb gadget hid device in nsprocon: %v", err)
	}
}

func (n *NSProCon) UpdateState(state *handler.GamepadState) error {
}

func (n *NSProCon) Press(buttons ...ButtonName) error {
}

func (n *NSProCon) Release(buttons ...ButtonName) error {
}

func (n *NSProCon) Push(buttons ...ButtonName, duration time.MilliSecond) error {
}

func (n *NSProCon) Repeat(buttons ...ButtonName, interval time.MilliSecond, duration time.MilliSecond) error {
}

func (n *NSProCon) StickL(xSir XDirection, xPower float64, yDir YDirection, yPower float64, duration time.MilliSecond) error {
}

func (n *NSProCon) StickR(xSir XDirection, xPower float64, yDir YDirection, yPower float64, duration time.MilliSecond) error {
}

func (n *NSProCon) StickRotationLeft(lapTime time.MilliSecond, power float64, duration time.MilliSecond) error {
}

func (n *NSProCon) StickRotationRight(lapTime time.MilliSecond, power float64, duration time.MilliSecond) error {
}

func NewNSProCon(verbose bool) *NSProcon {
	setupParams := &setup.UsbGadgetHidSetupParams{
		configsHome:     "/sys/kernel/config",
		gadgetName:      "nsprocon",
		idProduct:       "0x057e",
		idVendor:        "0x2009",
		bcdDevice:       "0x0200",
		bcdUsb:          "0x0200",
		bMaxPacketSize0: "64",
		bDeviceProtocol: "0",
		bDeviceSubClass: "0",
		bDeviceClass:    "0",
		stringsLang:     "0x409",
		iSerial:         "000000000001",
		iProduct:        "Pro Controller",
		iManufacture:    "Nintendo Co., Ltd.",
		configName:      "c",
		configNumber:    "1",
		configString:    "Nintendo Switch Pro Controller",
		bmAttributes:    "0xa0",
		maxPower:        "500mA",
		functionName:    "hid",
		instanceName:    setup.NSProConUGHIName,
		protocol:        "0",
		subclass:        "0",
		reportLength:    "203",
		reportDesc:      "050115000904A1018530050105091901290A150025017501950A5500650081020509190B290E150025017501950481027501950281030B01000100A1000B300001000B310001000B320001000B35000100150027FFFF0000751095048102C00B39000100150025073500463B0165147504950181020509190F2912150025017501950481027508953481030600FF852109017508953F8103858109027508953F8103850109037508953F9183851009047508953F9183858009057508953F9183858209067508953F9183C0",
	}
	return &NSProCon{
		verbose: verbose,
		setupParams: setupParams,
	}
}
