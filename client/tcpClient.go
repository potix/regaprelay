package client

import (
	"fmt"
	"log"
	"net"
	"github.com/potix/regapweb/handler"
	"github.com/potix/regaprelay/gamepad"
	"sync"
	"crypto/tls"
	"crypto/sha256"
	"bufio"
	"io"
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
	digest          string
	tlsConfig       *tls.Config
	gamepad         *gamepad.Gamepad
	connMutex       sync.Mutex
	conn            net.Conn
	stopCh          chan int
	remoteGamepadId string
	uid		string
	peerUid		string
}

func (t *TcpClient) safeConnWrite(msgBytes []byte) error  {
	t.connMutex.Lock()
	defer t.connMutex.Unlock()
	if t.conn == nil {
		log.Printf("no connection")
		return nil
	}
	 _, err := t.conn.Write(msgBytes)
	if err != nil {
               if err == io.EOF {
                        return nil
                } else {
			return fmt.Errorf("can not write message: %w", err)
		}
	}
	return nil
}

func (h *TcpClient) startPingLoop(conn net.Conn, pingLoopStopChan chan int) {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        for {
                select {
                case <-ticker.C:
                        msg := &handler.CommonMessage{
                                Command: "ping",
                        }
                        msgBytes, err := json.Marshal(msg)
                        if err != nil {
                                log.Printf("can not unmarshal to json: %v", err)
                                break
                        }
			msgBytes = append(msgBytes, byte('\n'))
                        _, err = conn.Write(msgBytes)
                        if err != nil {
				if err == io.EOF {
					return
				} else {
					log.Printf("can not write ping message: %v", err)
				}
                        }
                case <-pingLoopStopChan:
                        return
                }
        }
}

func (t *TcpClient) handshake(conn net.Conn) (string, error) {
	log.Printf("start handshake")
	var remoteGamepadId string
	msg := &handler.TcpClientRegisterRequest{
		CommonMessage: &handler.CommonMessage{
			Command: "registerRequest",
		},
		Digest: t.digest,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return remoteGamepadId, fmt.Errorf("can not marshal to json in handshake: %w", err)
	}
log.Printf("send msg = %v", string(msgBytes))
	msgBytes = append(msgBytes, byte('\n'))
	_, err = conn.Write(msgBytes)
	if err != nil {
		return remoteGamepadId, fmt.Errorf("can not write registerRequest: %w", err)
	}
        msgBytes = make([]byte, 0, 2048)
        rbufio := bufio.NewReader(conn)
	for {
		err := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			return remoteGamepadId, fmt.Errorf("can not set read deadline: %w", err)
		}
		patialMsgBytes, isPrefix, err := rbufio.ReadLine()
		if err != nil {
			return remoteGamepadId, fmt.Errorf("can not read message: %w", err)
		} else if isPrefix {
			// patial message
			msgBytes = append(msgBytes, patialMsgBytes...)
			continue
		} else {
			// entire message
			msgBytes = append(msgBytes, patialMsgBytes...)
log.Printf("recieve entire msg = %v", string(msgBytes))
			var msg handler.GamepadMessage
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				msgBytes = msgBytes[:0]
				return remoteGamepadId, fmt.Errorf("can not unmarshal message: %w", err)
			}
			msgBytes = msgBytes[:0]
			if msg.Command == "registerResponse" {
				if msg.RemoteGamepadId == "" {
					return remoteGamepadId, fmt.Errorf("no remote gamepad id in register response")
				}
				remoteGamepadId = msg.RemoteGamepadId
				return remoteGamepadId, nil
			} else {
				return remoteGamepadId, fmt.Errorf("recieved invalid message: %w", msg.Command)
			}
		}
	}
}

func (t *TcpClient) communicationLoop(conn net.Conn) error {
	log.Printf("start communication loop")
        remoteGamepadId, err := t.handshake(conn)
        if err != nil {
                return fmt.Errorf("can not handshakea: %w", err)
        }
	t.remoteGamepadId = remoteGamepadId
	log.Printf("remoteGamepadId = %v", t.remoteGamepadId)
	conn.SetDeadline(time.Time{})
        pingStopChan := make(chan int)
        go t.startPingLoop(conn, pingStopChan)
        defer close(pingStopChan)
	msgBytes := make([]byte, 0, 2048)
	rbufio := bufio.NewReader(conn)
	for {
		patialMsgBytes, isPrefix, err := rbufio.ReadLine()
		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				log.Printf("can not read message: %v", err)
				continue
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
				log.Printf("can not unmarshal message: %v, %v", string(msgBytes), err)
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
				var errMsg string
				if msg.RemoteGamepadId == t.remoteGamepadId {
					t.uid = msg.Uid
					t.peerUid = msg.PeerUid
					t.gamepad.UpdateState(msg.State)
				} else {
					errMsg = fmt.Sprintf("remote gamepad id is mismatch (act) %v, (exp) %v", msg.RemoteGamepadId, t.remoteGamepadId)
				}
				resMsg := &handler.GamepadMessage{
					CommonMessage: &handler.CommonMessage{
						Command: "stateResponse",
					},
					Error: errMsg,
				}
				resMsgBytes, err := json.Marshal(resMsg)
				if err != nil {
					log.Printf("can not marshal to json in communicationLoop: %v", err)
					continue
				}
				resMsgBytes = append(resMsgBytes, byte('\n'))
				_, err = conn.Write(resMsgBytes)
				if err != nil {
					if err == io.EOF {
                                                return nil
                                        } else {
						log.Printf("can not write state response message: %v", err)
					}
				}
			}
		}
	}
}

func (t *TcpClient) reconnectLoop() {
	log.Printf("start reconnect loop")
	for {
		select {
		case <-t.stopCh:
			log.Printf("stop reconnect loop")
		default:
		}
		log.Printf("connect to server %v", t.serverHostPort)
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
	}
}

func (t *TcpClient) onVibration(vibration *handler.GamepadVibrationMessage) {
	if t.uid == "" || t.peerUid == "" || t.remoteGamepadId == "" {
		return
	}
	msg := &handler.GamepadMessage {
	        CommonMessage: &handler.CommonMessage{
			Command: "vibrationRequest",
		},
		Uid: t.uid,
		PeerUid: t.peerUid,
		RemoteGamepadId: t.remoteGamepadId,
		Vibration: vibration,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("can not marshal to json in onVibration: %v", err)
		return
	}
	msgBytes = append(msgBytes, byte('\n'))
	err = t.safeConnWrite(msgBytes)
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

func NewTcpClient(serverHostPort string, secret string, gamepad *gamepad.Gamepad, opts ...TcpClientOption) (*TcpClient, error) {
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
		digest: digest,
                tlsConfig: conf,
                gamepad: gamepad,
		conn: nil,
		stopCh: make(chan int),
        }, nil
}

