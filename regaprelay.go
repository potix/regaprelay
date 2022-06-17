package main

import (
        "encoding/json"
        "flag"
        "github.com/potix/utils/signal"
        "github.com/potix/utils/configurator"
        "github.com/potix/regaprelay/gamepad"
        "github.com/potix/regaprelay/client"
        "github.com/potix/regaprelay/watcher"
        "log"
        "log/syslog"
)

type regaprelayTcpClientConfig struct {
        ServerHostPort string `toml:"serverHostPort"`
	Name           string `toml:"name`
	Secret         string `toml:"secret`
	SkipVerify     bool   `toml:"skipVerify`
}

type regaprelayGamepadConfig struct {
	Model       gamepad.GamepadModel `toml:model`
	MacAddr     string               `toml:macAddr`
	SpiMemory60 string               `toml:spiMemory60`
	SpiMemory80 string               `toml:spiMemory80`
	DevFilePath string               `toml:devFilePath`
	ConfigsHome string               `toml:configsHome`
	Udc         string               `toml:udc`
}

type regaprelayWatcherConfig struct {
	Enable         bool   `toml:enable`
	KeyboardDevice string `toml:keyboardDevice`
}

type regaprelayLogConfig struct {
        UseSyslog bool `toml:"useSyslog"`
}

type regaprelayConfig struct {
        Verbose   bool                       `toml:"verbose"`
        TcpClient *regaprelayTcpClientConfig `toml:"tcpClient"`
        Gamepad   *regaprelayGamepadConfig   `toml:"gamepad"`
        Watcher   *regaprelayWatcherConfig   `toml:"watcher"`
        Log       *regaprelayLogConfig       `toml:"log"`
}

type commandArguments struct {
        configFile string
}

func verboseLoadedConfig(config *regaprelayConfig) {
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

func main() {
        cmdArgs := new(commandArguments)
        flag.StringVar(&cmdArgs.configFile, "config", "./regapweb.conf", "config file")
        flag.Parse()
        cf, err := configurator.NewConfigurator(cmdArgs.configFile)
        if err != nil {
                log.Fatalf("can not create configurator: %v", err)
        }
        var conf regaprelayConfig
        err = cf.Load(&conf)
        if err != nil {
                log.Fatalf("can not load config: %v", err)
        }
        if conf.TcpClient == nil || conf.Gamepad == nil {
                log.Fatalf("invalid config")
        }
        if conf.Log != nil && conf.Log.UseSyslog {
                logger, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "aars")
                if err != nil {
                        log.Fatalf("can not create syslog: %v", err)
                }
                log.SetOutput(logger)
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
	// setup tcp client
        tcVerboseOpt := client.TcpClientVerbose(conf.Verbose)
	tcSkipVerify := client.TcpClientSkipVerify(conf.TcpClient.SkipVerify)
        newTcpClient, err := client.NewTcpClient(conf.TcpClient.ServerHostPort, conf.TcpClient.Name, conf.TcpClient.Secret, newGamepad, tcSkipVerify, tcVerboseOpt)
	if err != nil {
		log.Fatalf("can not create tcp client: %v", err)
	}
	var newKeyboardWatcher *watcher.KeyboardWatcher
	if conf.Watcher.Enable {
		// setup watcher
		kwVerboseOpt := watcher.KeyboardWatcherVerbose(conf.Verbose)
		newKeyboardWatcher, err = watcher.NewKeyboardWatcher(newGamepad, conf.Watcher.KeyboardDevice, watcher.ModeBulk, kwVerboseOpt)
		if err != nil {
			 log.Fatalf("can not create keyboard watcher: %v", err)
		}
	}
	err = newGamepad.Start()
	if err != nil {
		log.Fatalf("can not start gamepad: %v", err)
	}
	err = newTcpClient.Start()
	if err != nil {
		log.Fatalf("can not start tcp client: %v", err)
	}
	if newKeyboardWatcher != nil {
		newKeyboardWatcher.Start()
	}
        signal.SignalWait(nil)
	if newKeyboardWatcher != nil {
		newKeyboardWatcher.Stop()
	}
        newTcpClient.Stop()
        newGamepad.Stop()
}

