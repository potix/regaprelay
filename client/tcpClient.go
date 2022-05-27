package client

import (
	"log"
	"net"
	"github.com/potix/regapweb/handler"
	"sync"
	"crypto/tls"
	"bufio"
	"io"
	"encoding/json"
)

type tcpClientOptions struct {
        verbose    bool
	skipVerify bool
}

func defaultTcpClientOptions() *tcpClientOptions {
        return &tcpOptions {
                verbose: false,
                skipVerify: false,
        }
}

type TcpClientOption func(*tcpClientOptions)

func TcpClientVerbose(verbose bool) TcpClientOption {
        return func(opts *tcpClientOptions) {
                opts.verbose = verbose
        }
}

func TcpClientSkipVerify(skipVerify bool) TcpClientOption {
        return func(opts *tcpClientOptions) {
                opts.skipVerify = skipVerify
        }
}

type TcpClient struct {
	verbose        bool
	tlsConfig      tls.Config
	gamepad        *Gamepad
	serverAddrPort string
	connMutex      sync.Mutex
	conn           *net.Conn
	stopChan       chan int
}

func (t *TcpClient) safeConnWrite(msgBytes []byte) error  {
	t.connMutex.Lock()
	defer t.connMutex.Unlock()
	if t.conn == nil {
		log.Printf("no connection")
		return nil
	}
	 _, err = conn.Write(msgBytes)
	if err != nil {
		return fmt.Errorf("can not write message: %w", err)
	}
}

func (h *TcpClient) startPingLoop(conn *net.Conn, pingLoopStopChan chan int) {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        for {
                select {
                case <-ticker.C:
                        req := &handler.CommonGamepadMessage{
                                Command: "ping",
                        }
                        msgBytes, err := json.Marshal(req)
                        if err != nil {
                                log.Printf("can not unmarshal to json: %v", err)
                                break
                        }
                        err = conn.Write(conn, msgBytes)
                        if err != nil {
                                log.Printf("can not write ping message: %v", err)
                        }
                case <-pingLoopStopChan:
                        return
                }
        }
}

func (t *TcpClient) communicationLoop(conn *net.Conn) error {
        pingStopChan := make(chan int)
        go t.startPingLoop(conn, pingStopChan)
        defer close(pingStopChan)
	msgBytes := make([]byte, 0, 2048)
	rbufio := bufio.NewReader(conn)
	for {
		patialMsgBytes, isPrefix, err = rbufio.ReadLine()
		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				return fmt.Errorf("read error: %w", err)
			}
		} else if isPrefix {
			// patial message
			msgBytes = append(msgBytes, patialMsgBytes...)
			continue
		} else {
			// entire message
			msgBytes = append(msgBytes, patialMsgBytes...)
			var msg handler.GamepadMessage
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Printf("can not unmarshal message: %v", err)
				msgBytes = msgBytes[:0]
				continue
			}
			msgBytes = msgBytes[:0]
			if msg.Command == "ping" {
				continue
			} else if msg.Command == "vibrationResponse" {
				if msg.Error != "" {
					log.Printf("error vibration request: %v", msg.Error)
				}
			} else if msg.Command == "stateRequest" {
				t.gamepad.UpdateState(msg.GamepadState)
				resMsg := &gamepadMessage{
					Command: "stateResponse",
					Error: "",
				}
				resMsgBytes, err := json.Marshal(res)
				if err != nil {
					log.Printf("can not unmarshal to json in communicationLoop: %v", err)
					continue
				}
				err = conn.Write(conn, resMsgBytes)
				if err != nil {
					log.Printf("can not write state response message: %v", err)
				}
			}
		}
	}
}

func (t *TcpClient) reconnectLoop() error {
	for {
		select {
		case <-stopCh:
			log.Printf("stop reconnect loop")
			return
		}
		conn, err := tls.Dial("tcp", "127.0.0.1:4430", conf)
		if err != nil {
			log.Printf("can not connect to tcp server: %v", err)
			time.Sleep(500 * time.Millisecond)
			log.Printf("reconnect tcp server")
			continue
		}
		t.connMutex.Lock()
		t.conn = conn
		t.connMutex.Unlock()
		err = communicationLoop(conn)
		if err != nil {
			log.Printf("communication error: %v", err)
		}
		t.connMutex.Lock()
		t.conn = nil
		conn.Close()
		t.connMutex.Unlock()
	}
}

func (t *TcpHandler) onVibration(vibration *handler.GamepadVibration) {
	msg := &handler.GamepadMessage {
		Command: "vibrationRequest",
		Error: "",
		State: nil,
		Vibration: vibration,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("can not unmarshal to json in onVibration: %v", err)
		return
	}
	err = t.safeConnWrite(msgBytes)
	if err != nil {
		log.Printf("can not write vibration request message: %v", err)
	}
}

func (t *TcpClient) Start() error {
        go t.reconnectLoop()
	t.gamepad.StartVibrationListener(t.onVibration)
}

func (t *TcpClient) Stop() error {
	t.gamepad.StopVibrationListener()
	close(t.stopChan)
	t.connMutex.Lock()
	if t.conn != nil {
		t.conn.SetReadDeadline(time.Now())
	}
	t.connMutex.Unlock()
}

func NewTcpClient(serverAddrPort string, gamepad *Gamepad, opts ...TcpOption) (*TcpHandler, error) {
        baseOpts := defaultTcpClientOptions()
        for _, opt := range opts {
                if opt == nil {
                        continue
                }
                opt(baseOpts)
        }
	conf := &tls.Config{
		InsecureSkipVerify: baseOpts.skipVerify,
	}
	conn, err := tls.Dial("tcp", serverAddrPort, conf)
	if err != nil {
		return nil, fmt.Errorf("can not connect server: %w", err)
	}
	conn.Close()
        return &TcpClient{
                verbose: baseOpts.verbose,
                tlsConfig: conf,
                gamepad: gamepad,
		serverAddrPort: serverAddrPort,
		conn: *net.Conn,
		stopChan: make(chan int),
        }, nil
}

