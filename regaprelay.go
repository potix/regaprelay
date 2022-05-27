package main

import (
        "encoding/json"
        "flag"
        "github.com/potix/utils/signal"
        "github.com/potix/utils/configurator"
        "github.com/potix/regaprelay/gamepad"
        "github.com/potix/regaprelay/client"
        "log"
        "log/syslog"
)

type regaprelayTcpClientConfig struct {
        ServerHostPort string `toml:"serverHostPort"`
	SkipVerify     bool   `toml:"skipVerify`
}

type regaprelayGamepadConfig struct {
	Model gamepad.GamepadModel `toml:model`
}

type regaprelayLogConfig struct {
        UseSyslog bool `toml:"useSyslog"`
}

type regaprelayConfig struct {
        Verbose   bool                       `toml:"verbose"`
        TcpClient *regaprelayTcpClientConfig `toml:"tcpClient"`
        Gamepad   *regaprelayGamepadConfig   `toml:"gamepad"`
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
        newGamepad, err := gamepad.NewGamepad(conf.Gamepad.Model, gVerboseOpt)
	if err != nil {
		log.Fatalf("can not create gamepad: %v", err)
	}
	// setup tcp client
        tcVerboseOpt := client.TcpClientVerbose(conf.Verbose)
	tcSkipVerify := client.TcpClientSkipVerify(conf.TcpClient.SkipVerify)
        newTcpClient, err := client.NewTcpClient(conf.TcpClient.ServerHostPort, newGamepad, tcSkipVerify, tcVerboseOpt)
	if err != nil {
		log.Fatalf("can not create tcp client: %v", err)
	}
	err = newGamepad.Start()
	if err != nil {
		log.Fatalf("can not start gamepad: %v", err)
	}
	err = newTcpClient.Start()
	if err != nil {
		log.Fatalf("can not start tcp client: %v", err)
	}
        signal.SignalWait(nil)
        newTcpClient.Stop()
        newGamepad.Stop()
}

