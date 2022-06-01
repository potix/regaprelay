package watcher

import (
        "github.com/potix/regaprelay/gamepad"
	"github.com/azul3d/engine/keyboard"
        "log"
        "time"
)

type Mode string

const (
	Split mode = "Split"
	Bulk       = "Bulk"
)

type KeyboardWatcher struct {
	watcher      *keyboard.watcher
	gamepad      *gamepad.Gamepad
	mode         Mode
	buttonsState map[keyboard.Key]keyboard.State
	sticksState  map[keyboard.Key]keyboard.State
	shiftState   keyboard.State
	checkKeys    []keyboard.Key
	stickToggle  bool
	stopCh       chan int
}

func (k *keyboardWatcher) watch() {
	shiftChanged = false
	changed = false
	status := k.watcher.States()
	for _, key :=  range k.checkKeys {
		newState := status[key]
		if newState == keyboard.Up {
			if key == keyboard.LeftShift {
				// シフトが押された
				if l.shiftState == keyboard.Down {
					k.shiftState = newState
					shiftChanged = true
				}
			} else if k.stickToggle {
				oldState, ok = k.sticksState[key]
				if !ok {
					continue
				}
				// スティックが動かされた
				if oldState == keyboard.Down {
					k.sticksState[key] = newState
					changed = true
				}
			} else {
				oldState, ok = k.buttonsState[key]
				if !ok {
					continue
				}
				// ボタンが押された
				if oldState == keyboard.Down {
					k.buttonsState[key] = newState
					changed = true
				}
			}
		} else if s == keyboard.Down {
			if key == keyboard.LeftShift {
				// シフトが離された
				if l.shiftState == keyboard.Up {
					k.shiftState = newState
				}
			} else if k.stickToggle {
				oldState, ok = k.sticksState[key]
				if !ok {
					continue
				}
				// スティックが動きをとめた
				if oldState == keyboard.Up {
					k.sticksState[key] = newState
					changed = true
				}
			} else {
				oldState, ok = k.buttonsState[key]
				if !ok {
					continue
				}
				// ボタンが離された
				if oldState == keyboard.Up {
					k.buttonsState[key] = newState
					changed = true
				}
			}

		}
	}
	// shiftが押されていた場合の処理
	if shiftChanged {
		// stickToggleを切り替えるので片方は一回リセット
		if k.stickToggle {
			for _, key :=  range keys {
				_, ok = k.sticksState[key]
				if !ok {
					continue
				}
				k.sticksState[key] = keyboard.Up
			}
		} else {
			for _, key :=  range keys {
				_, ok = k.buttonsState[key]
				if !ok {
					continue
				}
				k.buttonsState[key] = keyboard.Up
			}
		}
		// stickToggle変更
		if k.stickToggle == false {
			k.stickToggle = false
		} else  {
			k.stickToggle = true
		}
	}
	// 何か変化があったらgamepadに送る
	if changed || shiftChanged {
		if k.mode == Bulk {
			buttons := make([]*GamepadButtonMessage, 17)
			keyOder := []keyboard.Key{
				 keyboard.K, // 0 : B 
				 keyboard.L, // 1 : A
				 keyboard.J, // 2 : Y
				 keyboard.I, // 3 : X
				 keyboard.F, // 4 : L
				 keyboard.H, // 5 : R
				 keyboard.E, // 6 : ZL
				 keyboard.U, // 7 : ZR
				 keyboard.C, // 8 : Minux
				 keyboard.N, // 9 : Plus
				 keyboard.Q, // 10: LStick
				 keyboard.P, // 11: RStick
				 keyboard.W, // 12: Up
				 keyboard.S, // 13: Down
				 keyboard.A, // 14: Left
				 keyboard.D, // 15: Right
				 keyboard.B, // 16: Home
				 keyboard.V, // 17: Capture
			}
			for i, key := keyOrder {
				state, ok = k.buttonsState[key]
				if !ok {
					log.Fatalf("not found key (%v,%v)", i, key)
				}
				if state == keyboard.Down {
					buttons[i] = &GamepadButtonMessage{
						Pressed: true,
						Touched: true,
						Value: 1.0,
					}
				} else {
					buttons[i] = &GamepadButtonMessage{
						Pressed: false,
						Touched: false,
						Value: 0.0,
					}
				}
			}
			axes := make([]float64, 4)
			for key, state := k.sticksState {
				switch key {
				case keyboard.W:
					if state == keyboard.Down {
						axes[1] += -1.0
					}
				case keyboard.S:
					if state == keyboard.Down {
						axes[1] += 1.0
					}
				case keyboard.A:
					if state == keyboard.Down {
						axes[0] += -1.0
					}
				case keyboard.D:
					if state == keyboard.Down {
						axes[0] += 1.0
					}
				case keyboard.I:
					if state == keyboard.Down {
						axes[3] += -1.0
					}
				case keyboard.K:
					if state == keyboard.Down {
						axes[3] += 1.0
					}
				case keyboard.J:
					if state == keyboard.Down {
						axes[2] += -1.0
					}
				case keyboard.L:
					if state == keyboard.Down {
						axes[2] += 1.0
					}
				}
			}
			gamepadStateMessage := &handler.GamepadStateMessage{
				Buttons: buttons,
				Axes: axes,
			}
			k.gamepad.UpdateState(gamepadStateMessage)
		} k.mode == Split {
			for key, state := k.buttonsState {
				switch key {
				case keyboard.W:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonUp)
					} else {
						k.gamepad.Release(gamepad.ButtonUp)
					}
				case keyboard.S:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonDown)
					} else {
						k.gamepad.Release(gamepad.ButtonDown)
					}
				case keyboard.A:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonLeft)
					} else {
						k.gamepad.Release(gamepad.ButtonLeft)
					}
				case keyboard.D:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonRight)
					} else {
						k.gamepad.Release(gamepad.ButtonRight)
					}
				case keyboard.I:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonX)
					} else {
						k.gamepad.Release(gamepad.ButtonX)
					}
				case keyboard.K:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonB)
					} else {
						k.gamepad.Release(gamepad.ButtonB)
					}
				case keyboard.J:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonY)
					} else {
						k.gamepad.Release(gamepad.ButtonY)
					}
				case keyboard.L:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonA)
					} else {
						k.gamepad.Release(gamepad.ButtonA)
					}
				case keyboard.F:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonL)
					} else {
						k.gamepad.Release(gamepad.ButtonL)
					}
				case keyboard.H:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonR)
					} else {
						k.gamepad.Release(gamepad.ButtonR)
					}
				case keyboard.E:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonZL)
					} else {
						k.gamepad.Release(gamepad.ButtonZL)
					}
				case keyboard.U:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonZR)
					} else {
						k.gamepad.Release(gamepad.ButtonZR)
					}
				case keyboard.C:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonMinus)
					} else {
						k.gamepad.Release(gamepad.ButtonMinus)
					}
				case keyboard.N:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonPlus)
					} else {
						k.gamepad.Release(gamepad.ButtonPlus)
					}
				case keyboard.V:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonCapture)
					} else {
						k.gamepad.Release(gamepad.ButtonCapture)
					}
				case keyboard.B:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonHome)
					} else {
						k.gamepad.Release(gamepad.ButtonHome)
					}
				case keyboard.Q:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonStickL)
					} else {
						k.gamepad.Release(gamepad.ButtonStickL)
					}
				case keyboard.P:
					if state == keyboard.Down {
						k.gamepad.Press(gamepad.ButtonStickR)
					} else {
						k.gamepad.Release(gamepad.ButtonStickR)
					}
				}
			}
			var lxAxis float64
			var lyAxis float64
			var rxAxis float64
			var ryAxis float64
			for key, state := k.sticksState {
				switch key {
				case keyboard.W:
					if state == keyboard.Down {
						lyAxis += -1.0
					}
				case keyboard.S:
					if state == keyboard.Down {
						lyAxis += 1.0
					}
				case keyboard.A:
					if state == keyboard.Down {
						lxAxis += -1.0
					}
				case keyboard.D:
					if state == keyboard.Down {
						lxAxis += 1.0
					}
				case keyboard.I:
					if state == keyboard.Down {
						ryAxis += -1.0
					}
				case keyboard.K:
					if state == keyboard.Down {
						ryAxis += 1.0
					}
				case keyboard.J:
					if state == keyboard.Down {
						rxAxis += -1.0
					}
				case keyboard.L:
					if state == keyboard.Down {
						rxAxis += 1.0
					}
				}
			}
			k.gamepad.StickL(lxAxis, lyAxis)
			k.gamepad.StickR(rxAxis, ryAxis)
		}
	}
}

func (k *keyboardWatcher) watchLoop() {
        ticker := time.NewTicker((time.Millisecond * 1000 / 60) + time.Millisecond)
        defer ticker.Stop()
        for {
                select {
                case <-ticker.C:
			k.watch()
                case <-k.stopCh:
                        return
                }
        }
}


func (k *keyboardWatcher) Start() {
	go k.watchLoop()
}

func (k *keyboardWatcher) Start() {
	close(k.stopCh)
}

func NewKeyboardWatcher(gamepad *gamepad.Gamepad, mode Mode) *KeyboardWatcher {
	keysState := map[keyboard.Key]keyboard.State{a
		keyboard.W: keybord.Up, // ButtonUp
		keyboard.S: keybord.Up, // ButtonDown
		keyboard.A: keybord.Up, // ButtonLeft
		keyboard.D: keybord.Up, // ButtonRight
		keyboard.I: keybord.Up, // ButtonX
		keyboard.K: keybord.Up, // ButtonB
		keyboard.J: keybord.Up, // ButtonY
		keyboard.L: keybord.Up, // ButtonA
		keyboard.F: keybord.Up, // ButtonL
		keyboard.H: keybord.Up, // ButtonR
		keyboard.E: keybord.Up, // ButtonZL
		keyboard.U: keybord.Up, // ButtonZR
		keyboard.C: keybord.Up, // ButtonMinus
		keyboard.N: keybord.Up, // ButtonPlus
		keyboard.V: keybord.Up, // ButtonCapture
		keyboard.B: keybord.Up, // ButtonHome
		keyboard.Q: keybord.Up, // ButtonStickL
		keyboard.P: keybord.Up, // ButtonStickR
	}
	sticksState := map[keyboard.Key]keyboard.State{a
		keyboard.W: keybord.Up, // left stick up
		keyboard.S: keybord.Up, // left stick down
		keyboard.A: keybord.Up, // left stick left
		keyboard.D: keybord.Up, // left stick right
		keyboard.I: keybord.Up, // right stick up
		keyboard.K: keybord.Up, // right stick down
		keyboard.J: keybord.Up, // right stick left
		keyboard.L: keybord.Up, // right stick right
	}
	checkKeys := make([]keyboard.Key, 0)
	for k, _ := range keyState {
		checkKeys = append(checkKeys, k)
	}
	checkKeys = append(checkKeys, keyboard.LeftShift)
	return &keyboardWatcher{
		watcher: keyboard.NewWatcher(),
		gamepad: gamepad,
		mode: mode,
		buttonsState: keysState,
		sticksState: sticksState,
		shiftState: keybord.Up,
		checkKeys checkKeys,
		stickToggle: false,
		stopCh: make(chan int),
	}
}



