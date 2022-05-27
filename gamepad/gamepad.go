package gamepad

import (
	"fmt"
	"time"
	"github.com/potix/regapweb/handler"
	"github.com/potix/regaprelay/backend"
)

type GamepadModel string

const (
        ModelNSProCon GamepadModel = "nsprocon"
        ModelPS4Con                = "ps4con"
)

type gamepadOptions struct {
        verbose bool
}

func defaultGamepadOptions() *gamepadOptions {
        return &gamepadOptions {
                verbose: false,
        }
}

type GamepadOption func(*gamepadOptions)

func GamepadVerbose(verbose bool) GamepadOption {
        return func(opts *gamepadOptions) {
                opts.verbose = verbose
        }
}

type Gamepad struct {
	verbose   bool
	opts	  *gamepadOptions
        backendIf backend.BackendIf
}

func (g *Gamepad) StartVibrationListener(fn backend.OnVibration) {
	g.backendIf.StartVibrationListener(fn)
}
func (g *Gamepad) StopVibrationListener() {
	g.backendIf.StopVibrationListener()
}

func (g *Gamepad) UpdateState(state *handler.GamepadState) error {
	return g.backendIf.UpdateState(state)
}

func (g *Gamepad) Press(buttons ...backend.ButtonName) error {
	return g.backendIf.Press(buttons)
}

func (g *Gamepad) Release(buttons ...backend.ButtonName) error {
	return g.backendIf.Release(buttons)
}

func (g *Gamepad) Push(duration time.Duration, buttons ...backend.ButtonName) error {
	return g.backendIf.Push(buttons, duration)
}

func (g *Gamepad) Repeat(interval time.Duration, duration time.Duration, buttons ...backend.ButtonName,) error {
	return g.backendIf.Repeat(buttons, interval, duration)
}

func (g *Gamepad) StickL(xDir backend.XDirection, xPower float64, yDir backend.YDirection, yPower float64, duration time.Duration) error {
	return g.backendIf.StickL(xDir, xPower, yDir, yPower, duration)
}

func (g *Gamepad) StickR(xDir backend.XDirection, xPower float64, yDir backend.YDirection, yPower float64, duration time.Duration) error {
	return g.backendIf.StickR(xDir, xPower, yDir, yPower, duration)
}

func (g *Gamepad) RotationStickL(xDir backend.XDirection, lapTime time.Duration, power float64, duration time.Duration) error {
	return g.backendIf.RotationStickL(xDir, lapTime, power, duration)
}

func (g *Gamepad) RotationStickR(xDir backend.XDirection, lapTime time.Duration, power float64, duration time.Duration) error {
	return g.backendIf.RotationStickR(xDir, lapTime, power, duration)
}

func (g *Gamepad) Start() error {
	return g.backendIf.Start()
}

func (g *Gamepad) Stop() {
	g.backendIf.Stop()
}

func NewGamepad(model GamepadModel, opts ...GamepadOption) (*Gamepad, error) {
        baseOpts := defaultGamepadOptions()
        for _, opt := range opts {
                if opt == nil {
                        continue
                }
                opt(baseOpts)
        }
	var newBackendIf backend.BackendIf
	if model == ModelNSProCon {
		newBackendIf = backend.NewNSProCon(baseOpts.verbose)
	} else if model == ModelPS4Con {
		newBackendIf = backend.NewPS4Con(baseOpts.verbose)
	}
	if newBackendIf == nil {
		return nil, fmt.Errorf("unsupported model: %v", model)
	}
	err := newBackendIf.Setup()
	if err != nil {
		return nil, fmt.Errorf("backend setup error: %w", err)
	}
        return &Gamepad{
                verbose: baseOpts.verbose,
                backendIf: newBackendIf,
        }, nil
}
