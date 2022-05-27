package backend

import(
	"fmt"
	"log"
	"time"
	"github.com/potix/regaprelay/backend/setup"
	"github.com/potix/regapweb/handler"
)

type NSProCon struct  {
	BaseBackend
	setupParams *setup.UsbGadgetHidSetupParams
	verbose     bool
}

func (n *NSProCon) Setup() error {
	err := setup.UsbGadgetHidSetup(n.setupParams)
	if err != nil {
		return fmt.Errorf("setup error in nsprocon: %w", err)
	}
	return nil
}

func (n *NSProCon) Start() error {
	err := setup.UsbGadgetHidEnable(n.setupParams)
	if err != nil {
		return fmt.Errorf("can not enable usb gadget hid device in nsprocon: %w", err)
	}
	return nil
}

func (n *NSProCon) Stop() {
	err := setup.UsbGadgetHidDisable(n.setupParams)
	if err != nil {
		log.Printf("can not enable usb gadget hid device in nsprocon: %v", err)
	}
}

func (n *NSProCon) UpdateState(state *handler.GamepadState) error {
	// XXX TODO
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

func NewNSProCon(verbose bool) *NSProCon {
	setupParams := &setup.UsbGadgetHidSetupParams{
		ConfigsHome:     "/sys/kernel/config",
		GadgetName:      "nsprocon",
		IdProduct:       "0x057e",
		IdVendor:        "0x2009",
		BcdDevice:       "0x0200",
		BcdUsb:          "0x0200",
		BMaxPacketSize0: "64",
		BDeviceProtocol: "0",
		BDeviceSubClass: "0",
		BDeviceClass:    "0",
		StringsLang:     "0x409",
		ISerial:         "000000000001",
		IProduct:        "Pro Controller",
		IManufacture:    "Nintendo Co., Ltd.",
		ConfigName:      "c",
		ConfigNumber:    "1",
		ConfigString:    "Nintendo Switch Pro Controller",
		BmAttributes:    "0xa0",
		MaxPower:        "500mA",
		FunctionName:    "hid",
		InstanceName:    "usb0",
		Protocol:        "0",
		Subclass:        "0",
		ReportLength:    "203",
		ReportDesc:      "050115000904A1018530050105091901290A150025017501950A5500650081020509190B290E150025017501950481027501950281030B01000100A1000B300001000B310001000B320001000B35000100150027FFFF0000751095048102C00B39000100150025073500463B0165147504950181020509190F2912150025017501950481027508953481030600FF852109017508953F8103858109027508953F8103850109037508953F9183851009047508953F9183858009057508953F9183858209067508953F9183C0",
	}
	return &NSProCon{
		verbose: verbose,
		setupParams: setupParams,
	}
}
