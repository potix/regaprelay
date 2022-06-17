package client

import (
	"fmt"
	"log"
	"net"
	"github.com/potix/regapweb/message"
	"github.com/potix/regaprelay/gamepad"
	"sync"
	"crypto/tls"
	"crypto/sha256"
	"bufio"
	"time"
	"encoding/json"
)

type tcpClientOptions struct {
        verbose    bool
	skipVerify bool
}

func defaultTcpClientOptions() *tcpClientOptions {
        return &tcpClientOptions {
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
	verbose         bool
	serverHostPort  string
	name            string
	digest          string
	tlsConfig       *tls.Config
	gamepad         *gamepad.Gamepad
	connMutex       sync.Mutex
	conn            net.Conn
	stopCh          chan int
	gamepadId	string
	delivererId	string
	controllerId	string
}

func (t *TcpClient) safeConnWriteMessage(msg *message.Message) error  {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("can not marshal to json: %v", err)
	}
	msgBytes = append(msgBytes, byte('\n'))
	t.connMutex.Lock()
	defer t.connMutex.Unlock()
	if t.conn == nil {
		log.Printf("no connection")
		return nil
	}
	 _, err = t.conn.Write(msgBytes)
	if err != nil {
		return fmt.Errorf("can not write message: %w", err)
	}
	return nil
}

func (t *TcpClient) writeMessage(conn net.Conn, msg *message.Message) error  {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("can not marshal to json: %w", err)
	}
	msgBytes = append(msgBytes, byte('\n'))
	 _, err = conn.Write(msgBytes)
	if err != nil {
		return fmt.Errorf("can not write message: %w", err)
	}
	return nil
}

func (t *TcpClient) startPingLoop(conn net.Conn, pingLoopStopChan chan int) {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        for {
                select {
                case <-ticker.C:
                        msg := &message.Message{
                                MsgType: message.MsgTypePing,
                        }
			err := t.writeMessage(conn, msg)
                        if err != nil {
				log.Printf("can not write ping message: %v", err)
				return
                        }
                case <-pingLoopStopChan:
                        return
                }
        }
}

func (t *TcpClient) handshake(conn net.Conn) (string, error) {
	var gamepadId string
	msg := &message.Message{
		MsgType: message.MsgTypeGamepadHandshakeReq,
		GamepadHandshakeRequest: &message.GamepadHandshakeRequest {
			Name: t.name,
			Digest: t.digest,
		},
	}
	err := t.writeMessage(conn, msg)
	if err != nil {
		return gamepadId, fmt.Errorf("can not write gpHandshakeReq: %w", err)
	}
        msgBytes := make([]byte, 0, 4096)
        rbufio := bufio.NewReader(conn)
	for {
		err := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			return gamepadId, fmt.Errorf("can not set read deadline: %w", err)
		}
		patialMsgBytes, isPrefix, err := rbufio.ReadLine()
		if err != nil {
			return gamepadId, fmt.Errorf("can not read message: %w", err)
		} else if isPrefix {
			// patial message
			msgBytes = append(msgBytes, patialMsgBytes...)
			continue
		} else {
			// entire message
			msgBytes = append(msgBytes, patialMsgBytes...)
			var msg message.Message
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				msgBytes = msgBytes[:0]
				return gamepadId, fmt.Errorf("can not unmarshal message: %w", err)
			}
			msgBytes = msgBytes[:0]
			if msg.MsgType == message.MsgTypeGamepadHandshakeRes {
				if msg.GamepadHandshakeResponse == nil ||
				   msg.GamepadHandshakeResponse.GamepadId == "" {
					return gamepadId, fmt.Errorf("no gamepad id in handshake response")
				}
				gamepadId = msg.GamepadHandshakeResponse.GamepadId
				return gamepadId, nil
			} else {
				return gamepadId, fmt.Errorf("recieved invalid message: %w", msg.MsgType)
			}
		}
	}
}

func (t *TcpClient) communicationLoop(conn net.Conn) error {
	if t.verbose {
		log.Printf("start handshake")
	}
        gamepadId, err := t.handshake(conn)
        if err != nil {
                return fmt.Errorf("can not handshakea: %w", err)
        }
	if t.verbose {
		log.Printf("end handshake")
	}
	t.gamepadId = gamepadId
	log.Printf("gamepadId = %v", t.gamepadId)
	conn.SetDeadline(time.Time{})
        pingStopChan := make(chan int)
        go t.startPingLoop(conn, pingStopChan)
        defer close(pingStopChan)
	msgBytes := make([]byte, 0, 2048)
	rbufio := bufio.NewReader(conn)
	for {
		patialMsgBytes, isPrefix, err := rbufio.ReadLine()
		if err != nil {
			return fmt.Errorf("can not read message: %v", err)
		} else if isPrefix {
			// patial message
			msgBytes = append(msgBytes, patialMsgBytes...)
			continue
		} else {
			// entire message
			msgBytes = append(msgBytes, patialMsgBytes...)
			var msg message.Message
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Printf("can not unmarshal message: %v, %v", string(msgBytes), err)
				msgBytes = msgBytes[:0]
				continue
			}
			msgBytes = msgBytes[:0]
			if msg.MsgType == message.MsgTypePing {
				if t.verbose {
					log.Printf("recieved ping")
				}
				continue
			} else if msg.MsgType == message.MsgTypeGamepadConnectReq {
				if msg.GamepadConnectRequest == nil ||
				   msg.GamepadConnectRequest.DelivererId == "" ||
				   msg.GamepadConnectRequest.ControllerId == "" ||
				   msg.GamepadConnectRequest.GamepadId == "" {
					log.Printf("no gamepad connect request parameter: %v", msg.GamepadConnectRequest)
					resMsg := &message.Message{
						MsgType: message.MsgTypeGamepadConnectRes,
						Error: &message.Error{
							Message: "no gamepad connect request parameter",
						},
						GamepadConnectResponse: &message.GamepadConnectResponse{
							DelivererId: msg.GamepadConnectRequest.DelivererId,
							ControllerId: msg.GamepadConnectRequest.ControllerId,
							GamepadId: msg.GamepadConnectRequest.GamepadId,
						},
					}
					err = t.writeMessage(conn, resMsg)
					if err != nil {
						log.Printf("can not write gamepad connect response: %v", err)
						return fmt.Errorf("can not write gamepad connect response: %w", err)
					}
					continue
				}
				t.delivererId = msg.GamepadConnectRequest.DelivererId
				t.controllerId = msg.GamepadConnectRequest.ControllerId
				t.gamepadId = msg.GamepadConnectRequest.GamepadId
				resMsg := &message.Message{
					MsgType: message.MsgTypeGamepadConnectRes,
					GamepadConnectResponse: &message.GamepadConnectResponse{
						DelivererId: msg.GamepadConnectRequest.DelivererId,
						ControllerId: msg.GamepadConnectRequest.ControllerId,
						GamepadId: msg.GamepadConnectRequest.GamepadId,
					},
				}
				err = t.writeMessage(conn, resMsg)
				if err != nil {
					log.Printf("can not write gamepad connect response: %v", err)
					return fmt.Errorf("can not write gamepad connect response: %w", err)
				}
			} else if msg.MsgType == message.MsgTypeGamepadConnectServerError {
				if msg.Error != nil && msg.Error.Message != "" {
					log.Printf("error has occured in gpConnectRes: %v", msg.Error.Message)
				}
			} else if msg.MsgType == message.MsgTypeGamepadState {
				if msg.GamepadState == nil ||
				   msg.GamepadState.DelivererId == "" ||
				   msg.GamepadState.ControllerId == "" ||
				   msg.GamepadState.GamepadId == "" {
					log.Printf("no gamepad state request parameter: %v", msg.GamepadState)
					continue
				}
				if msg.GamepadState.GamepadId != t.gamepadId ||
				   msg.GamepadState.DelivererId != t.delivererId ||
				   msg.GamepadState.ControllerId != t.controllerId {
					log.Printf("ids are mismatch: gamepadId: (act) %v, (exp) %v, delivererId: (act) %v, (exp) %v, controllerId: (act) %v, (exp) %v",
						 msg.GamepadState.GamepadId, t.gamepadId, msg.GamepadState.DelivererId, t.delivererId, msg.GamepadState.ControllerId, t.controllerId)
					continue
				}
				t.gamepad.UpdateState(msg.GamepadState)
			} else {
				log.Printf("unsupported message: %v", msg.MsgType)
			}
		}
	}
}

func (t *TcpClient) reconnectLoop() {
	if t.verbose {
		log.Printf("start reconnect loop")
	}
	for {
		select {
		case <-t.stopCh:
			if t.verbose {
				log.Printf("stop reconnect loop")
			}
			break
		default:
		}
		if t.verbose {
			log.Printf("connect to server %v", t.serverHostPort)
		}
		conn, err := tls.Dial("tcp", t.serverHostPort, t.tlsConfig)
		if err != nil {
			log.Printf("can not connect to tcp server: %v", err)
			time.Sleep(500 * time.Millisecond)
			log.Printf("reconnect tcp server")
			continue
		}
		t.connMutex.Lock()
		t.conn = conn
		t.connMutex.Unlock()
		err = t.communicationLoop(conn)
		if err != nil {
			log.Printf("communication error: %v", err)
		}
		t.connMutex.Lock()
		t.conn = nil
		conn.Close()
		t.connMutex.Unlock()
		time.Sleep(500 * time.Millisecond)
	}
	if t.verbose {
		log.Printf("finish reconnect loop")
	}
}

func (t *TcpClient) onVibration(vibration *message.GamepadVibration) {
	if t.delivererId == "" || t.controllerId == "" || t.gamepadId == "" {
		if t.verbose {
			log.Printf("skip vibration because no ids")
		}
		return
	}
	vibration.DelivererId = t.delivererId
	vibration.ControllerId = t.controllerId
	vibration.GamepadId = t.gamepadId
	msg := &message.Message {
		MsgType: message.MsgTypeGamepadVibration,
		GamepadVibration: vibration,
	}
	err := t.safeConnWriteMessage(msg)
	if err != nil {
		log.Printf("can not write vibration request message: %v", err)
	}
}

func (t *TcpClient) Start() error {
        go t.reconnectLoop()
	t.gamepad.StartVibrationListener(t.onVibration)
	return nil
}

func (t *TcpClient) Stop() {
	t.gamepad.StopVibrationListener()
	close(t.stopCh)
	t.connMutex.Lock()
	if t.conn != nil {
		t.conn.SetReadDeadline(time.Now())
	}
	t.connMutex.Unlock()
}

func NewTcpClient(serverHostPort string, name string, secret string, gamepad *gamepad.Gamepad, opts ...TcpClientOption) (*TcpClient, error) {
        baseOpts := defaultTcpClientOptions()
        for _, opt := range opts {
                if opt == nil {
                        continue
                }
                opt(baseOpts)
        }
        sha := sha256.New()
        digest := fmt.Sprintf("%x", sha.Sum([]byte(secret)))
	conf := &tls.Config{
		InsecureSkipVerify: baseOpts.skipVerify,
	}
	conn, err := tls.Dial("tcp", serverHostPort, conf)
	if err != nil {
		return nil, fmt.Errorf("can not connect server: %w", err)
	}
	conn.Close()
        return &TcpClient{
                verbose: baseOpts.verbose,
		serverHostPort: serverHostPort,
		name: name,
		digest: digest,
                tlsConfig: conf,
                gamepad: gamepad,
		conn: nil,
		stopCh: make(chan int),
        }, nil
}

