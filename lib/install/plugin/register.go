// Copyright 2016 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/vic/pkg/trace"
	"github.com/vmware/vic/pkg/vsphere/session"

	"golang.org/x/net/context"
)

type Info struct {
	Company               string
	Key                   string
	Name                  string
	ShowInSolutionManager bool
	Summary               string
	Type                  string
	URL                   string
	Version               string
}

type Pluginator struct {
	Session          *session.Session
	ExtensionManager *object.ExtensionManager
	Context          context.Context

	info *Info

	tURL      *url.URL
	connected bool
}

func NewPluginator(ctx context.Context, target *url.URL, i *Info) (*Pluginator, error) {
	defer trace.End(trace.Begin(""))

	p := &Pluginator{
		tURL: target,
		info: i,
	}
	p.Context = ctx

	err := p.connect()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Pluginator) disconnect() error {
	defer trace.End(trace.Begin(""))
	if p.Session != nil {
		err := p.Session.Client.Logout(p.Context)
		if err != nil {
			return fmt.Errorf("failed to disconnect: %s", err)
		}
	}
	p.connected = false
	return nil
}

func (p *Pluginator) connect() error {
	defer trace.End(trace.Begin(""))
	var err error

	sessionconfig := &session.Config{
		Insecure: true,
	}
	sessionconfig.Service = p.tURL.String()

	p.Session = session.NewSession(sessionconfig)
	p.Session, err = p.Session.Connect(p.Context)
	if err != nil {
		return fmt.Errorf("failed to connect: %s", err)
	}

	p.Session.Populate(p.Context)

	em, err := object.GetExtensionManager(p.Session.Client.Client)
	if err != nil {
		return fmt.Errorf("failed to get extension manager: %s", err)
	}
	p.ExtensionManager = em

	p.connected = true
	return nil
}

// Register installs an extension to the target
func (p *Pluginator) Register() error {
	defer trace.End(trace.Begin(""))
	var err error
	if !p.connected {
		return errors.New("not connected")
	}

	desc := types.Description{
		Label:   p.info.Name,
		Summary: p.info.Summary,
	}

	e := types.Extension{
		Key:         p.info.Key,
		Version:     p.info.Version,
		Company:     p.info.Company,
		Description: &desc,
	}

	eci := types.ExtensionClientInfo{
		Version:     p.info.Version,
		Company:     p.info.Company,
		Description: &desc,
		Type:        p.info.Type,
		Url:         p.info.URL,
	}
	e.Client = append(e.Client, eci)

	d := types.KeyValue{
		Key:   "name",
		Value: p.info.Name,
	}

	eri := types.ExtensionResourceInfo{
		Locale: "en_US",
		Module: "name",
	}

	eri.Data = append(eri.Data, d)

	e.ResourceList = append(e.ResourceList, eri)

	// TODO HTTP server info

	e.ShownInSolutionManager = &p.info.ShowInSolutionManager

	e.LastHeartbeatTime = time.Now().UTC()

	err = p.ExtensionManager.Register(p.Context, e)
	if err != nil {
		return err
	}

	return nil
}

// Unregister removes an extension from the target
func (p *Pluginator) Unregister(key string) error {
	defer trace.End(trace.Begin(""))
	if !p.connected {
		return errors.New("not connected")
	}

	if err := p.ExtensionManager.Unregister(p.Context, key); err != nil {
		return err
	}
	return nil
}

// IsRegistered checks for presence of an extension on the target
func (p *Pluginator) IsRegistered(key string) (bool, error) {
	defer trace.End(trace.Begin(""))
	if !p.connected {
		return false, errors.New("not connected")
	}

	e, err := p.ExtensionManager.Find(p.Context, key)
	if err != nil {
		return false, err
	}
	if e != nil {
		log.Debugf("%q is registered", key)
		return true, nil
	}
	log.Debugf("%q is not registered", key)
	return false, nil
}
