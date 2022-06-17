package gamepad

//
// XXX TODO
// I don't have a PlayStation 4.
// I want a PlayStation 5 rather than a PlayStation 4.
// Perhaps I will never implement this in my lifetime.
//

import(
        "fmt"
        "log"
	"github.com/potix/regaprelay/gamepad/setup"
	"github.com/potix/regapweb/message"
)

type PS4Con struct  {
	*BaseBackend
	setupParams *setup.UsbGadgetHidSetupParams
	verbose     bool
	devFilePath string
}

func (p *PS4Con) Setup() error {
        err := setup.UsbGadgetHidSetup(p.setupParams)
        if err != nil {
                return fmt.Errorf("setup error in ps4con: %w", err)
        }
        return nil
}

func (p *PS4Con) Start() error {
        err := setup.UsbGadgetHidEnable(p.setupParams)
        if err != nil {
                return fmt.Errorf("can not enable usb gadget hid device in nsprocon: %w", err)
        }
        return nil
}

func (p *PS4Con) Stop() {
        err := setup.UsbGadgetHidDisable(p.setupParams)
        if err != nil {
                log.Printf("can not enable usb gadget hid device in nsprocon: %v", err)
        }
}

func (p *PS4Con) UpdateState(state *message.GamepadState) error {
	// XXX TODO
        return nil
}

func (p *PS4Con) Press(buttons []ButtonName) error {
	// XXX TODO
        return nil
}

func (p *PS4Con) Release(buttons []ButtonName) error {
	// XXX TODO
        return nil
}

func (p *PS4Con) StickL(xAxis float64, yAxis float64) error {
	// XXX TODO
        return nil
}

func (p *PS4Con) StickR(xAxis float64, yAxis float64) error {
	// XXX TODO
        return nil
}

func NewPS4Con(verbose bool, devFilePath string, configsHome string, udc string) *PS4Con {
        setupParams := &setup.UsbGadgetHidSetupParams{
                ConfigsHome:     configsHome,
                GadgetName:      "ps4con",
                IdProduct:       "0x05c4",
                IdVendor:        "0x054c",
                BcdDevice:       "0x0100",
                BcdUsb:          "0x0200",
                BMaxPacketSize0: "64",
                BDeviceProtocol: "0",
                BDeviceSubClass: "0",
                BDeviceClass:    "0",
                StringsLang:     "0x409",
                ISerial:         "0",
                IProduct:        "Wireless Controller",
                IManufacturer:   "Sony Computer Entertainment",
                ConfigName:      "c",
                ConfigNumber:    "1",
                ConfigString:    "Play Station 4 Controller",
                BmAttributes:    "0xc0",
                MaxPower:        "500",
                FunctionName:    "hid",
                InstanceName:    "usb0",
                Protocol:        "0",
                Subclass:        "0",
                ReportLength:    "499",
                ReportDesc:      "05010905A10185010930093109320935150026FF007508950481020939150025073500463B016514750495018142650005091901290E150025017501950E81020600FF0920750695011500257F8102050109330934150026FF007508950281020600FF09219536810285050922951F9102850409239524B102850209249524B102850809259503B102851009269504B102851109279502B10285120602FF0921950FB102851309229516B10285140605FF09209510B10285150921952CB1020680FF858009209506B102858109219506B102858209229505B102858309239501B102858409249504B102858509259506B102858609269506B102858709279523B102858809289522B102858909299502B102859009309505B102859109319503B102859209329503B10285930933950CB10285A009409506B10285A109419501B10285A209429501B10285A309439530B10285A40944950DB10285A509459515B10285A609469515B10285F00947953FB10285F10948953FB10285F20949950FB10285A7094A9501B10285A8094B9501B10285A9094C9508B10285AA094E9501B10285AB094F9539B10285AC09509539B10285AD0951950BB10285AE09529501B10285AF09539502B10285B00954953FB10285B109559502B10285B209569502B10285B30955953FB10285B40955953FB102C0",
		UDC:             udc,
        }
	defaultDevFilePath := "XXX"
	if devFilePath == "" {
		devFilePath = defaultDevFilePath
	}
	return &PS4Con{
		BaseBackend: &BaseBackend{},
		setupParams: setupParams,
		verbose: verbose,
		devFilePath: devFilePath,
	}
}
