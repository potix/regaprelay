package gamepad

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
	XDirectionNeutral
	XDirectionLeft
	XDirectionRgiht
)

type YDirection int

const (
	YDirectionNeutral
	YDirectionUP
	YDirectionDown
)

type BackendIf interface {
	Setup() error
	Start() error
	Stop()
	UpdateState(*handler.GamepadState) error
	Press(...ButtonName) error
        Release(...ButtonName) error
        Push(...ButtonName, time.MilliSecond) error
	Repeat(...ButtonName, time.MilliSecond, time.MilliSecond) error
        StickL(XDirection, float64, YDirection, float64, time.MilliSecond) error
	StickR(XDirection, float64, YDirection, float64, time.MilliSecond) error
	StickRotationLeft(time.MilliSecond, float64, time.MilliSecond) error
        StickRotationRight(time.MilliSecond, float64, time.MilliSecon) error
	StartVibrationListener(fn onVibration)
	StopVibrationListener()
}

type BaseBackend struct {
	onVibrationChan         chan *handler.GamepadVibration
	stopVibrationListenerCh chan int
}

type OnVibration func(*handler.GamepadVibration)

func (b *BaseBackend) StartVibrationListener(fn onVibration) {
	b.onVibrationCh = make(chan *handler.GamepadVibration)
	b.stopVibrationListenerCh: = make(chan int)
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
	if b.onVibrationChan != nil {
		b.onVibrationCh <- vibration
	}
}
