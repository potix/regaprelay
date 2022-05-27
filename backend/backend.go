package gamepad

import (
	"log"
	"time"
	"github.com/potix/regapweb/handler"
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

type XDirection int

const (
	XDirectionNeutral XDirection = iota
	XDirectionLeft
	XDirectionRgiht
)

type YDirection int

const (
	YDirectionNeutral YDirection = iota
	YDirectionUP
	YDirectionDown
)

type OnVibration func(*handler.GamepadVibration)

type BackendIf interface {
	Setup() error
	Start() error
	Stop()
	UpdateState(*handler.GamepadState) error
	Press(...ButtonName) error
        Release(...ButtonName) error
        Push(time.Duration, ...ButtonName) error
	Repeat(time.Duration, time.Duration, ...ButtonName) error
        StickL(XDirection, float64, YDirection, float64, time.Duration) error
	StickR(XDirection, float64, YDirection, float64, time.Duration) error
	StickRotationLeft(time.Duration, float64, time.Duration) error
        StickRotationRight(time.Duration, float64, time.Duration) error
	StartVibrationListener(fn OnVibration)
	StopVibrationListener()
}

type BaseBackend struct {
	onVibrationCh           chan *handler.GamepadVibration
	stopVibrationListenerCh chan int
}

func (b *BaseBackend) StartVibrationListener(fn OnVibration) {
	b.onVibrationCh = make(chan *handler.GamepadVibration)
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

func (b *BaseBackend) sendVibration(vibration *handler.GamepadVibration) {
	if b.onVibrationCh != nil {
		b.onVibrationCh <- vibration
	}
}
