package backend

import (
	"log"
	"time"
	"github.com/potix/regapweb/handler"
	"github.com/potix/regaprelay/gamepad"
)

type OnVibration func(*handler.GamepadVibrationMessage)

type BackendIf interface {
	Setup() error
	Start() error
	Stop()
	UpdateState(*handler.GamepadStateMessage) error
	Press([]gamepad.ButtonName) error
        Release([]gamepad.ButtonName) error
        StickL(float64, float64) error
	StickR(float64, float64) error
	StartVibrationListener(fn OnVibration)
	StopVibrationListener()
}

type BaseBackend struct {
	onVibrationCh           chan *handler.GamepadVibrationMessage
	stopVibrationListenerCh chan int
}

func (b *BaseBackend) StartVibrationListener(fn OnVibration) {
	b.onVibrationCh = make(chan *handler.GamepadVibrationMessage)
	b.stopVibrationListenerCh = make(chan int)
        go func() {
                log.Printf("start vibration listener")
                for {
                        select {
                        case v := <-b.onVibrationCh:
                                fn(v)
                        case <-b.stopVibrationListenerCh:
                                return
                        }
                }
                log.Printf("finish vibration listener")
        }()
}

func (b *BaseBackend) StopVibrationListener() {
	if b.stopVibrationListenerCh != nil {
		close(b.stopVibrationListenerCh)
	}
}

func (b *BaseBackend) sendVibration(vibration *handler.GamepadVibrationMessage) {
	if b.onVibrationCh != nil {
		b.onVibrationCh <- vibration
	}
}
