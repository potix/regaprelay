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
	configsHome     string
	gadgetName      string
	idProduct       string
	idVendor        string
	bcdDevice       string
	bcdUsb          string
	bMaxPacketSize0 string
	bDeviceProtocol string
	bDeviceSubClass string
	bDeviceClass    string
	stringsLang     string
	iSerial         string
	iProduct        string
	iManufacture    string
	configName      string
	configNumber    string
	configString    string
	bmAttributes    string
	maxPower        string
	functionName    string
	instanceName    string
	protocol        string
	subclass        string
	reportLength    string
	reportDesc      string
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
