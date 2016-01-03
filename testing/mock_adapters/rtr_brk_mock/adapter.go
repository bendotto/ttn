// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package rtr_brk_mock offers a router <-> broker mock adapter that can be used to test a router
// implementation.
package rtr_brk_mock

import (
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"time"
)

type Adapter struct {
	Brokers       []core.BrokerAddress                           // All known brokers with which the router is communicating
	FailListen    bool                                           // If true, any call to Listen will fail with an error
	FailBroadcast bool                                           // If true, any call to Broadcast will trigger a core.ErrBroadcast
	FailForward   bool                                           // If true, any call to Forward will trigger a core.ErrForward
	Broadcasts    map[semtech.DeviceAddress][]semtech.Payload    // Stores all payload send through broadcasts
	Forwards      map[semtech.DeviceAddress][]semtech.Payload    // Stores all payload send through forwards
	Relations     map[semtech.DeviceAddress][]core.BrokerAddress // Known association between brokers and device
}

// New constructs a new router <-> broker adapter interface
func New() *Adapter {
	return &Adapter{
		FailListen:    false,
		FailBroadcast: false,
		FailForward:   false,
		Broadcasts:    make(map[semtech.DeviceAddress][]semtech.Payload),
		Forwards:      make(map[semtech.DeviceAddress][]semtech.Payload),
		Relations:     make(map[semtech.DeviceAddress][]core.BrokerAddress),
	}
}

// Connect implements the core.Adapter interface. Expect a slice of broker address as options
func (a *Adapter) Listen(router core.Router, options interface{}) error {
	if a.FailListen {
		return core.ErrBadOptions
	}
	a.Brokers = options.([]core.BrokerAddress)
	return nil
}

// Broadcast implements the core.BrokerRouter interface
func (a *Adapter) Broadcast(router core.Router, payload semtech.Payload) error {
	devAddr, err := payload.UniformDevAddr()

	if a.FailBroadcast || payload.RXPK == nil || err != nil {
		return core.ErrBroadcast
	}

	<-time.After(time.Millisecond * 50)
	a.Broadcasts[*devAddr] = append(a.Broadcasts[*devAddr], payload)
	brokers, ok := a.Relations[*devAddr]
	if ok {
		router.RegisterDevice(*devAddr, brokers...)
	}
	return nil
}

// Forward implements the core.BrokerRouter interface
func (a *Adapter) Forward(router core.Router, payload semtech.Payload, broAddrs ...core.BrokerAddress) error {
	devAddr, err := payload.UniformDevAddr()

	if a.FailForward || err != nil {
		return core.ErrForward
	}
	a.Forwards[*devAddr] = append(a.Forwards[*devAddr], payload)
	return nil
}
