package setup

import (

)

// ==============================
// get usb hid device information
// ==============================
// - get usb list
// sudo lsusb
// - get usb detail
// sudo lsusb -d <idProduct>:<idVendor> -v 2> /dev/null
// - get report descriptor
// sudo usbhid-dump -d <idProduct>:<idVendor>

const usbGadgetDir = "usb_gadget"
const usbDevCon    = "UDC"

type UsbGadgetHidSetupParams struct {
	ConfigsHome     string
	GadgetName      string
	IdProduct       string
	IdVendor        string
	BcdDevice       string
	BcdUsb          string
	BMaxPacketSize0 string
	BDeviceProtocol string
	BDeviceSubClass string
	BDeviceClass    string
	StringsLang     string
	ISerial         string
	IProduct        string
	IManufacture    string
	ConfigName      string
	ConfigNumber    string
	ConfigString    string
	BmAttributes    string
	MaxPower        string
	FunctionName    string
	InstanceName    string
	Protocol        string
	Subclass        string
	ReportLength    string
	ReportDesc      string
}

func UsbGadgetHidSetup(params *UsbGadgetHidSetupParams) error {
	return nil
}

func UsbGadgetHidEnable(params *UsbGadgetHidSetupParams) error {
	return nil
}

func UsbGadgetHidDisable(params *UsbGadgetHidSetupParams) error {
	return nil
}
