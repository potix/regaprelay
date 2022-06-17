package gamepad

import (
	"log"
	"github.com/potix/regapweb/message"
)

type GamepadModel string

const (
        ModelNSProCon GamepadModel = "nsprocon"
        ModelPS4Con                = "ps4con"
)

type ButtonName int

const (
        ButtonA ButtonName = iota
        ButtonB
        ButtonX
        ButtonY
        ButtonLeft
        ButtonRight
        ButtonUp
        ButtonDown
        ButtonPlus
        ButtonMinus
        ButtonHome
        ButtonCapture
        ButtonStickL
        ButtonStickR
        ButtonL
        ButtonR
        ButtonZL
        ButtonZR
        ButtonLeftSL
        ButtonLeftSR
        ButtonRightSL
        ButtonRightSR
        ButtonChargingGrip
)

type BackendIf interface {
	Setup() error
	Start() error
	Stop()
	UpdateState(*message.GamepadState) error
	Press([]ButtonName) error
        Release([]ButtonName) error
        StickL(float64, float64) error
	StickR(float64, float64) error
	StartVibrationListener(fn OnVibration)
	StopVibrationListener()
}

type BaseBackend struct {
	verbose			bool
	onVibrationCh           chan *message.GamepadVibration
	stopVibrationListenerCh chan int
}

func (b *BaseBackend) StartVibrationListener(fn OnVibration) {
	b.onVibrationCh = make(chan *message.GamepadVibration)
	b.stopVibrationListenerCh = make(chan int)
        go func() {
		if b.verbose {
			log.Printf("start vibration listener")
		}
                for {
                        select {
                        case v := <-b.onVibrationCh:
                                fn(v)
                        case <-b.stopVibrationListenerCh:
                                return
                        }
                }
		if b.verbose {
			log.Printf("finish vibration listener")
		}
        }()
}

func (b *BaseBackend) StopVibrationListener() {
	if b.stopVibrationListenerCh != nil {
		close(b.stopVibrationListenerCh)
	}
}

func (b *BaseBackend) SendVibration(vibration *message.GamepadVibration) {
	if b.onVibrationCh != nil {
		b.onVibrationCh <- vibration
	}
}
