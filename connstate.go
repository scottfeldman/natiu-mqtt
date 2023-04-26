package mqtt

import (
	"errors"
	"io"
	"sync"
	"time"
)

type clientState struct {
	mu          sync.Mutex
	lastRx      time.Time
	connectedAt time.Time
	pendingSubs []SubscribeRequest
	activeSubs  []string
	// closeErr stores the reason for disconnection.
	closeErr error
}

// callbacks returns the Rx and Tx callbacks necessary for a clientState to function automatically.
// The onPub callback
func (cs *clientState) callbacks(onPub func(rx *Rx, varPub VariablesPublish, r io.Reader) error) (RxCallbacks, TxCallbacks) {
	closeConn := func(err error) {
		cs.mu.Lock()
		defer cs.mu.Unlock()
		cs.connectedAt = time.Time{}
		cs.closeErr = err
	}
	return RxCallbacks{
			OnConnack: func(r *Rx, vc VariablesConnack) error {
				connTime := time.Now()
				cs.mu.Lock()
				defer cs.mu.Unlock()
				cs.lastRx = connTime
				if vc.ReturnCode != 0 {
					return errors.New(vc.ReturnCode.String())
				}
				cs.connectedAt = connTime
				return nil
			},
			OnPub: onPub,
			OnSuback: func(r *Rx, vs VariablesSuback) error {
				rxTime := time.Now()
				cs.mu.Lock()
				defer cs.mu.Unlock()
				cs.lastRx = rxTime
				if len(vs.ReturnCodes) != len(cs.pendingSubs) {
					return errors.New("got mismatched number of return codes compared to pending client subscriptions")
				}
				for i, qos := range vs.ReturnCodes {
					if qos != QoSSubfail {
						if qos != cs.pendingSubs[i].QoS {
							return errors.New("QoS does not match requested QoS for topic")
						}
						cs.activeSubs = append(cs.activeSubs, string(cs.pendingSubs[i].TopicFilter))
					}
				}
				cs.pendingSubs = cs.pendingSubs[:0]
				return nil
			},
			OnOther: func(rx *Rx, packetIdentifier uint16) (err error) {
				tp := rx.LastReceivedHeader.Type()
				rxTime := time.Now()
				cs.mu.Lock()
				defer cs.mu.Unlock()
				cs.lastRx = rxTime
				switch tp {
				case PacketDisconnect:
					cs.connectedAt = time.Time{}
					err = errors.New("received graceful disconnect request")
				case PacketPingresp:

				}
				if err != nil {
					cs.closeErr = err
				}
				return err
			},
			OnRxError: func(r *Rx, err error) {
				closeConn(err)
			},
			// OnOther: ,
		}, TxCallbacks{
			OnTxError: func(tx *Tx, err error) {
				closeConn(err)
			},
		}
}

// Connected returns true if the client is currently connected.
func (cs *clientState) Connected() bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if cs.connectedAt.IsZero() == (cs.closeErr != nil) {
		panic("assertion failed: bug in natiu-mqtt clientState implementation")
	}
	return cs.closeErr == nil
}

// Err returns the error that caused the MQTT connection to finish.
// Returns nil if currently connected.
func (cs *clientState) Err() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if cs.connectedAt.IsZero() == (cs.closeErr != nil) {
		panic("assertion failed: bug in natiu-mqtt clientState implementation")
	}
	return cs.closeErr
}

// PendingResponse returns true if the client is waiting on the server for a response.
func (cs *clientState) PendingResponse() bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.closeErr == nil && len(cs.pendingSubs) > 0
}
