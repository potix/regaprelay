package main

import (
        "encoding/json"
        "flag"
	"github.com/potix/regapweb/handler"
        "github.com/potix/utils/signal"
        "github.com/potix/utils/configurator"
        "github.com/potix/regaprelay/gamepad"
        "github.com/potix/regaprelay/watcher"
	"github.com/azul3d/engine/keyboard"
        "log"
)

}

func newKeyboardWatcher() *keyboardWatcher {
	return &keyboardWatcher{
		watcher: keyboard.NewWatcher(),
		stopCh: make(chan int),
	}
}








type gpadtestGamepadConfig struct {
	Model       gamepad.GamepadModel `toml:model`
	MacAddr     string               `toml:macAddr`
	SpiMemory60 string               `toml:spiMemory60`
	SpiMemory80 string               `toml:spiMemory80`
	DevFilePath string               `toml:devFilePath`
	ConfigsHome string               `toml:configsHome`
	Udc         string               `toml:udc`
}

type gpadtestConfig struct {
        Verbose   bool                     `toml:"verbose"`
        Gamepad   *gpadtestGamepadConfig   `toml:"gamepad"`
}

type commandArguments struct {
        configFile string
        mode string
}

func verboseLoadedConfig(config *gpadtestConfig) {
        if !config.Verbose {
                return
        }
        j, err := json.Marshal(config)
        if err != nil {
                log.Printf("can not dump config: %v", err)
                return
        }
        log.Printf("loaded config: %v", string(j))
}

func onVibration(vibration *handler.GamepadVibrationMessage) {
	log.Printf("get vibration -> %v", vibration)
}

func main() {
        cmdArgs := new(commandArguments)
        flag.StringVar(&cmdArgs.configFile, "config", "./gpadtest.conf", "config file")
        flag.StringVar(&cmdArgs.mode, "mode", "bulk", "mode('bulk' or 'split')")
        flag.Parse()
        cf, err := configurator.NewConfigurator(cmdArgs.configFile)
        if err != nil {
                log.Fatalf("can not create configurator: %v", err)
        }
        var conf gpadtestConfig
        err = cf.Load(&conf)
        if err != nil {
                log.Fatalf("can not load config: %v", err)
        }
        if conf.Gamepad == nil {
                log.Fatalf("invalid config")
        }
        verboseLoadedConfig(&conf)
	// setup gamepad
        gVerboseOpt := gamepad.GamepadVerbose(conf.Verbose)
        gDevFilePathOpt := gamepad.GamepadDevFilePath(conf.Gamepad.DevFilePath)
        gConfigsHomeOpt := gamepad.GamepadConfigsHome(conf.Gamepad.ConfigsHome)
        gUdcOpt := gamepad.GamepadUdc(conf.Gamepad.Udc)
        newGamepad, err := gamepad.NewGamepad(conf.Gamepad.Model, conf.Gamepad.MacAddr, conf.Gamepad.SpiMemory60, conf.Gamepad.SpiMemory80, gDevFilePathOpt, gConfigsHomeOpt, gUdcOpt, gVerboseOpt)
	if err != nil {
		log.Fatalf("can not create gamepad: %v", err)
	}
        newGamepad.StartVibrationListener(onVibration)
	// setup watcher
	newKeyboardWatcher := watcher.NewKeyboardWatcher(newGamepad, cmdArgs.mode)
	// start gamepad
	err = newGamepad.Start()
	if err != nil {
		log.Fatalf("can not start gamepad: %v", err)
	}
	// start keyboardWatcher
	newKeyboardWatcher.Start()
        signal.SignalWait(nil)
	newKeyboardWatcher.Stop()
        newGamepad.Stop()
}

