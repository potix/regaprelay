package gamepad

import (
	"fmt"
	"github.com/potix/regapweb/handler"
)

type gamepadOptions struct {
        verbose     bool
	devFilePath string
	configsHome string
	udc         string
}

func defaultGamepadOptions() *gamepadOptions {
        return &gamepadOptions {
                verbose: false,
		devFilePath: "",
		configsHome: "",
		udc: "",
        }
}

type GamepadOption func(*gamepadOptions)

func GamepadVerbose(verbose bool) GamepadOption {
        return func(opts *gamepadOptions) {
                opts.verbose = verbose
        }
}

func GamepadDevFilePath(devFilePath string) GamepadOption {
        return func(opts *gamepadOptions) {
                opts.devFilePath = devFilePath
        }
}

func GamepadConfigsHome(configsHome string) GamepadOption {
        return func(opts *gamepadOptions) {
                opts.configsHome = configsHome
        }
}

func GamepadUdc(udc string) GamepadOption {
        return func(opts *gamepadOptions) {
                opts.udc = udc
        }
}

type Gamepad struct {
	verbose   bool
	opts	  *gamepadOptions
        backendIf BackendIf
}

func (g *Gamepad) StartVibrationListener(fn OnVibration) {
	g.backendIf.StartVibrationListener(fn)
}
func (g *Gamepad) StopVibrationListener() {
	g.backendIf.StopVibrationListener()
}

func (g *Gamepad) UpdateState(state *handler.GamepadStateMessage) error {
	return g.backendIf.UpdateState(state)
}

func (g *Gamepad) Press(buttons ...ButtonName) error {
	return g.backendIf.Press(buttons)
}

func (g *Gamepad) Release(buttons ...ButtonName) error {
	return g.backendIf.Release(buttons)
}

func (g *Gamepad) StickL(xAxis float64, yAxis float64) error {
	return g.backendIf.StickL(xAxis, yAxis)
}

func (g *Gamepad) StickR(xAxis float64, yAxis float64) error {
	return g.backendIf.StickR(xAxis, yAxis)
}

func (g *Gamepad) Start() error {
	return g.backendIf.Start()
}

func (g *Gamepad) Stop() {
	g.backendIf.Stop()
}

func NewGamepad(model GamepadModel, macAddr string, spiMemory60 string, spiMemory80 string, opts ...GamepadOption) (*Gamepad, error) {
        baseOpts := defaultGamepadOptions()
        for _, opt := range opts {
                if opt == nil {
                        continue
                }
                opt(baseOpts)
        }
	var err error
	var newBackendIf BackendIf
	if model == ModelNSProCon {
		newBackendIf, err = NewNSProCon(baseOpts.verbose, macAddr, spiMemory60, spiMemory80, baseOpts.devFilePath, baseOpts.configsHome, baseOpts.udc)
		if err != nil {
			return nil, fmt.Errorf("can not create procon: %v", err)
		}
	} else if model == ModelPS4Con {
		newBackendIf = NewPS4Con(baseOpts.verbose, baseOpts.devFilePath, baseOpts.configsHome, baseOpts.udc)
	}
	if newBackendIf == nil {
		return nil, fmt.Errorf("unsupported model: %v", model)
	}
	err = newBackendIf.Setup()
	if err != nil {
		return nil, fmt.Errorf("backend setup error: %w", err)
	}
        return &Gamepad{
                verbose: baseOpts.verbose,
                backendIf: newBackendIf,
        }, nil
}
