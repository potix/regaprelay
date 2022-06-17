package watcher

import (
        "github.com/potix/regaprelay/gamepad"
	"github.com/potix/regapweb/message"
	"github.com/MarinX/keylogger"
        "log"
        "fmt"
)

type Mode string

const (
	ModeSplit Mode = "Split"
	ModeBulk       = "Bulk"
)

type keyboardWatcherOptions struct {
        verbose    bool
}

func defaultKeyboardWatcherOptions() *keyboardWatcherOptions {
        return &keyboardWatcherOptions {
                verbose: false,
        }
}

type KeyboardWatcherOption func(*keyboardWatcherOptions)

func KeyboardWatcherVerbose(verbose bool) KeyboardWatcherOption {
        return func(opts *keyboardWatcherOptions) {
                opts.verbose = verbose
        }
}

type KeyboardWatcher struct {
	verbose	            bool
	keyLogger           *keylogger.KeyLogger
	gamepad             *gamepad.Gamepad
	mode                Mode
	gamepadButtonsOrder []string
	gamepadButtonsMap   map[string]gamepad.ButtonName
	buttonsState        map[string]bool
	sticksState         map[string]bool
	shiftState          bool
	checkKeys           []string
	stickToggle         bool
	stopCh              chan int
}

func (k *KeyboardWatcher) updateState(event keylogger.InputEvent) {
	shiftChanged := false
	changed := false
	switch event.Type {
	case keylogger.EvKey:
		key := event.KeyString()
		if event.KeyPress() {
			if k.verbose {
				log.Printf("[event] press key %v", key)
			}
			if key == "L_SHIFT" {
				// シフトが押された
				k.shiftState = true
				shiftChanged = true
			} else if k.stickToggle {
				_, ok := k.sticksState[key]
				if !ok {
					break
				}
				// スティックが動かされた
				k.sticksState[key] = true
				changed = true
			} else {
				_, ok := k.buttonsState[key]
				if !ok {
					break
				}
				// ボタンが押された
				k.buttonsState[key] = true
				changed = true
			}
		} else if event.KeyRelease() {
			if k.verbose {
				log.Printf("[event] release key %v", key)
			}
			if key == "L_SHIFT" {
				// シフトが離された
				k.shiftState = false
			} else if k.stickToggle {
				_, ok := k.sticksState[key]
				if !ok {
					break
				}
				// スティックが動きをとめた
				k.sticksState[key] = false
				changed = true
			} else {
				_, ok := k.buttonsState[key]
				if !ok {
					break
				}
				// ボタンが離された
				k.buttonsState[key] = false
				changed = true
			}
		}
	}
	// shiftが押されていた場合の処理
	if shiftChanged {
		// stickToggleを切り替えるので片方は一回リセット
		if k.stickToggle {
			for key, _ := range k.sticksState {
				k.sticksState[key] = false
			}
		} else {
			for key, _ :=  range k.buttonsState {
				k.buttonsState[key] = false
			}
		}
		// stickToggle変更
		if k.stickToggle {
			k.stickToggle = false
		} else  {
			k.stickToggle = true
		}
	}
	// 何か変化があったらgamepadに送る
	if changed || shiftChanged {
		if k.mode == ModeBulk {
			buttons := make([]*message.GamepadButtonState, 18)
			for i, key := range k.gamepadButtonsOrder {
				pressed, ok := k.buttonsState[key]
				if !ok {
					log.Fatalf("not found key (%v,%v)", i, key)
				}
				if pressed {
					buttons[i] = &message.GamepadButtonState{
						Pressed: true,
						Touched: true,
						Value: 1.0,
					}
				} else {
					buttons[i] = &message.GamepadButtonState{
						Pressed: false,
						Touched: false,
						Value: 0.0,
					}
				}
			}
			axes := make([]float64, 4)
			for key, pressed := range k.sticksState {
				switch key {
				case "W":
					if pressed {
						axes[1] += -1.0
					}
				case "S":
					if pressed {
						axes[1] += 1.0
					}
				case "A":
					if pressed {
						axes[0] += -1.0
					}
				case "D":
					if pressed {
						axes[0] += 1.0
					}
				case "I":
					if pressed {
						axes[3] += -1.0
					}
				case "K":
					if pressed {
						axes[3] += 1.0
					}
				case "J":
					if pressed {
						axes[2] += -1.0
					}
				case "L":
					if pressed {
						axes[2] += 1.0
					}
				}
			}
			gamepadState := &message.GamepadState{
				Buttons: buttons,
				Axes: axes,
			}
			k.gamepad.UpdateState(gamepadState)
		} else if k.mode == ModeSplit {
			for key, pressed := range k.buttonsState {
				buttonName, ok := k.gamepadButtonsMap[key]
				if !ok {
					log.Fatalf("not found key (%v)", key)
				}
				if pressed {
					k.gamepad.Press(buttonName)
				} else {
					k.gamepad.Release(buttonName)
				}
			}
			var lxAxis float64
			var lyAxis float64
			var rxAxis float64
			var ryAxis float64
			for key, pressed := range k.sticksState {
				switch key {
				case "W":
					if pressed {
						lyAxis += -1.0
					}
				case "S":
					if pressed {
						lyAxis += 1.0
					}
				case "A":
					if pressed {
						lxAxis += -1.0
					}
				case "D":
					if pressed {
						lxAxis += 1.0
					}
				case "I":
					if pressed {
						ryAxis += -1.0
					}
				case "K":
					if pressed {
						ryAxis += 1.0
					}
				case "J":
					if pressed {
						rxAxis += -1.0
					}
				case "L":
					if pressed {
						rxAxis += 1.0
					}
				}
			}
			k.gamepad.StickL(lxAxis, lyAxis)
			k.gamepad.StickR(rxAxis, ryAxis)
		}
	}
}

func (k *KeyboardWatcher) watchLoop() {
	inputEventCh := k.keyLogger.Read()
	for event := range inputEventCh {
		 k.updateState(event)
	}
}

func (k *KeyboardWatcher) Start() {
	go k.watchLoop()
}

func (k *KeyboardWatcher) Stop() {
	k.keyLogger.Close()
}

func NewKeyboardWatcher(gpad *gamepad.Gamepad, keyboardDevice string, mode Mode, opts ...KeyboardWatcherOption) (*KeyboardWatcher, error) {
        baseOpts := defaultKeyboardWatcherOptions()
        for _, opt := range opts {
                if opt == nil {
                        continue
                }
                opt(baseOpts)
        }
	gamepadButtonsOrder := []string{
		 "K", // 0 : B 
		 "L", // 1 : A
		 "J", // 2 : Y
		 "I", // 3 : X
		 "F", // 4 : L
		 "H", // 5 : R
		 "E", // 6 : ZL
		 "U", // 7 : ZR
		 "C", // 8 : Minux
		 "N", // 9 : Plus
		 "Q", // 10: LStick
		 "O", // 11: RStick
		 "W", // 12: Up
		 "S", // 13: Down
		 "A", // 14: Left
		 "D", // 15: Right
		 "B", // 16: Home
		 "V", // 17: Capture
	}
	gamepadButtonsMap := map[string]gamepad.ButtonName {
		"W": gamepad.ButtonUp,
		"S": gamepad.ButtonDown,
		"A": gamepad.ButtonLeft,
		"D": gamepad.ButtonRight,
		"I": gamepad.ButtonX,
		"K": gamepad.ButtonB,
		"J": gamepad.ButtonY,
		"L": gamepad.ButtonA,
		"F": gamepad.ButtonL,
		"H": gamepad.ButtonR,
		"E": gamepad.ButtonZL,
		"U": gamepad.ButtonZR,
		"C": gamepad.ButtonMinus,
		"N": gamepad.ButtonPlus,
		"V": gamepad.ButtonCapture,
		"B": gamepad.ButtonHome,
		"Q": gamepad.ButtonStickL,
		"O": gamepad.ButtonStickR,
	}
	buttonsState := map[string]bool{
		"W": false, // ButtonUp
		"S": false, // ButtonDown
		"A": false, // ButtonLeft
		"D": false, // ButtonRight
		"I": false, // ButtonX
		"K": false, // ButtonB
		"J": false, // ButtonY
		"L": false, // ButtonA
		"F": false, // ButtonL
		"H": false, // ButtonR
		"E": false, // ButtonZL
		"U": false, // ButtonZR
		"C": false, // ButtonMinus
		"N": false, // ButtonPlus
		"V": false, // ButtonCapture
		"B": false, // ButtonHome
		"Q": false, // ButtonStickL
		"O": false, // ButtonStickR
	}
	sticksState := map[string]bool{
		"W": false, // left stick up
		"S": false, // left stick down
		"A": false, // left stick left
		"D": false, // left stick right
		"I": false, // right stick up
		"K": false, // right stick down
		"J": false, // right stick left
		"L": false, // right stick right
	}
	checkKeys := make([]string, 0)
	for k, _ := range buttonsState {
		checkKeys = append(checkKeys, k)
	}
	checkKeys = append(checkKeys, "L_SHIFT")
	if keyboardDevice == "" {
		keyboardDevice = keylogger.FindKeyboardDevice()
	}
	newKeylogger, err := keylogger.New(keyboardDevice)
	if err != nil {
		return nil, fmt.Errorf("can not create key logger: %w", err)
	}
	return &KeyboardWatcher{
		verbose: baseOpts.verbose,
		keyLogger: newKeylogger,
		gamepad: gpad,
		mode: mode,
		gamepadButtonsOrder: gamepadButtonsOrder,
		gamepadButtonsMap: gamepadButtonsMap,
		buttonsState: buttonsState,
		sticksState: sticksState,
		shiftState: false,
		checkKeys: checkKeys,
		stickToggle: false,
		stopCh: make(chan int),
	}, nil
}



