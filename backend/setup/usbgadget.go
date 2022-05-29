package setup

import (
	"os"
	"path"
	"fmt"
	"encoding/hex"
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
	IManufacturer   string
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


func remove(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		err = os.Remove(path)
		if err != nil {
			return fmt.Errorf("can not remove (%v): %w", path, err)
		}
	}
	return nil
}

func makeDir(dir string, perm os.FileMode) error {
	err := os.Mkdir(dir, perm)
	if err != nil &&  !os.IsExist(err)  {
		return fmt.Errorf("can not create gadget dir (%v): %w", dir, err)
	}
	return nil
}

func makeDirAll(dir string, perm os.FileMode) error {
	err := os.MkdirAll(dir, perm)
	if err != nil {
		return fmt.Errorf("can not create gadget dir (%v): %w", dir, err)
	}
	return nil
}

func writeToFile(dirName string, fileName string, value string, perm os.FileMode) error {
	filePath := path.Join(dirName, fileName)
	return os.WriteFile(filePath, []byte(value), perm)
}

func writeHexStringToFile(dirName string, fileName string, hexString string, perm os.FileMode) error {
	decodedBytes, err := hex.DecodeString(hexString)
	if err != nil {
		return fmt.Errorf("can not decode hex string (%v): %w", hexString, err)
	}
	filePath := path.Join(dirName, fileName)
	return os.WriteFile(filePath, decodedBytes, perm)
}

func createSymlink(oldName string, newName string) error {
	_, err := os.Stat(newName)
	if err == nil {
		err = os.Remove(newName)
		if err != nil {
			return fmt.Errorf("can not remove already exists symbolic link (%v): %w", newName, err)
		}
	}
	err = os.Symlink(oldName, newName)
	if err != nil {
		return fmt.Errorf("can not create new symbolic link (%v -> %v): %w", newName, oldName, err)
	}
	return nil
}

func UsbGadgetHidSetup(params *UsbGadgetHidSetupParams) error {
	// e.g. /sys/kernel/config/usb_gadget/<name>
	gadgetDir := path.Join(params.ConfigsHome, usbGadgetDir, params.GadgetName)
	// e.g. /sys/kernel/config/usb_gadget/<name>/strings/0x409
	stringsDir := path.Join(gadgetDir, "strings", params.StringsLang)
	// e.g. /sys/kernel/config/usb_gadget/<name>/configs/c.1
	configsDir := path.Join(gadgetDir, "configs", params.ConfigName + "." + params.ConfigNumber)
	// e.g. /sys/kernel/config/usb_gadget/<name>/configs/c.1/strings/0x409
	configsStringsDir := path.Join(configsDir, "strings", params.StringsLang)
	// e.g. /sys/kernel/config/usb_gadget/<name>/functions/hid.usb0
	functionsDir := path.Join(gadgetDir, "functions", params.FunctionName + "." + params.InstanceName)
	// e.g. /sys/kernel/config/usb_gadget/<name>/configs/c.1/hid.usb0
	configsFunctionDir := path.Join(configsDir, params.FunctionName + "." + params.InstanceName)

	// cleanup
	err := remove(configsFunctionDir)
	if err != nil {
		return fmt.Errorf("can not remove config function symlink (%v): %w", configsFunctionDir, err)
	}
	err = remove(configsStringsDir)
	if err != nil {
		return fmt.Errorf("can not remove strings dir in configs dir (%v): %w", configsStringsDir, err)
	}
	err = remove(configsDir)
	if err != nil {
		return fmt.Errorf("can not remove configs dir (%v): %w", configsDir, err)
	}
	err = remove(functionsDir)
	if err != nil {
		return fmt.Errorf("can not remove functions dir (%v): %w", functionsDir, err)
	}
	err = remove(stringsDir)
	if err != nil {
		return fmt.Errorf("can not remove stringss dir (%v): %w", stringsDir, err)
	}
	err = remove(gadgetDir)
	if err != nil {
		return fmt.Errorf("can not remove gadget dir (%v): %w", gadgetDir, err)
	}
	// setup /sys/kernel/config/usb_gadget/<name>/*
	err = makeDir(gadgetDir, 0755)
	if err != nil {
		return fmt.Errorf("can not create gadget dir (%v): %w", gadgetDir, err)
	}
	err = writeToFile(gadgetDir, "idVendor", params.IdVendor, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "idVendor", err)
	}
	err = writeToFile(gadgetDir, "idProduct", params.IdProduct, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "idProduct", err)
	}
	err = writeToFile(gadgetDir, "bcdDevice", params.BcdDevice, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "bcdDevice", err)
	}
	err = writeToFile(gadgetDir, "bcdUSB", params.BcdUsb, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "bcdUSB", err)
	}
	err = writeToFile(gadgetDir, "bMaxPacketSize0", params.BMaxPacketSize0, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "bMaxPacketSize0", err)
	}
	err = writeToFile(gadgetDir, "bDeviceClass", params.BDeviceClass, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "bDeviceClass", err)
	}
	err = writeToFile(gadgetDir, "bDeviceSubClass", params.BDeviceSubClass, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "bDeviceSubClass", err)
	}
	err = writeToFile(gadgetDir, "bDeviceProtocol", params.BDeviceProtocol, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "bDeviceProtocol", err)
	}
	// setup  /sys/kernel/config/usb_gadget/<name>/strings/0x409/*
	err = makeDirAll(stringsDir, 0755)
	if err != nil {
		return fmt.Errorf("can not create strings dir (%v): %w", stringsDir, err)
	}
	err = writeToFile(stringsDir, "serialnumber", params.ISerial, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "serialnumber", err)
	}
	err = writeToFile(stringsDir, "manufacturer", params.IManufacturer, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "manufacturer", err)
	}
	err = writeToFile(stringsDir, "product", params.IProduct, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "product", err)
	}
	// setup /sys/kernel/config/usb_gadget/<name>/configs/c.1/*
	err = makeDirAll(configsDir, 0755)
	if err != nil {
		return fmt.Errorf("can not create configs dir (%v): %w", configsDir, err)
	}
	err = writeToFile(configsDir, "bmAttributes", params.BmAttributes, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "bmAttributes", err)
	}
	err = writeToFile(configsDir, "MaxPower", params.MaxPower, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "MaxPower", err)
	}
	// setup /sys/kernel/config/usb_gadget/<name>/configs/c.1/strings/0x409/*
	err = makeDirAll(configsStringsDir, 0755)
	if err != nil {
		return fmt.Errorf("can not create strings dir in configs dir (%v): %w", configsStringsDir, err)
	}
	err = writeToFile(configsStringsDir, "configuration", params.ConfigString, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "configuration", err)
	}
	// setup /sys/kernel/config/usb_gadget/<name>/functions/hid.usb0
	err = makeDirAll(functionsDir, 0755)
	if err != nil {
		return fmt.Errorf("can not create functions dir (%v): %w", functionsDir, err)
	}
	err = writeToFile(functionsDir, "protocol", params.Protocol, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "protocol", err)
	}
	err = writeToFile(functionsDir, "subclass", params.Subclass, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "subclass", err)
	}
	err = writeToFile(functionsDir, "report_length", params.ReportLength, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "report_length", err)
	}
	err = writeHexStringToFile(functionsDir, "report_desc", params.ReportDesc, 0644)
	if err != nil {
		return fmt.Errorf("can not write to file (%v): %w", "report_desc", err)
	}
	// setup /sys/kernel/config/usb_gadget/<name>/configs/c.1/hid.usb0 -> /sys/kernel/config/usb_gadget/<name>/functions/hid.usb0
	err = os.Symlink(functionsDir, configsFunctionDir)
	if err != nil {
		return fmt.Errorf("can not create symbolic link (%v -> %v): %w", configsFunctionDir, functionsDir, err)
	}
	return nil
}

func UsbGadgetHidEnable(params *UsbGadgetHidSetupParams) error {
	return nil
}

func UsbGadgetHidDisable(params *UsbGadgetHidSetupParams) error {
	return nil
}
