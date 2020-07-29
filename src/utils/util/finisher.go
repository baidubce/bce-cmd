// Copyright 2017 Baidu, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
// except in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the
// License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific language governing permissions
// and limitations under the License.

package util

import (
	"sync"
)

var (
	GFinisher *Finisher
)

func init() {
	// global Finisher
	GFinisher = &Finisher{
		queue: make(map[string]CompleteFuncInter),
	}
}

type CompleteFuncInter interface {
	Exit() error
	GetId() (string, error)
}

type Finisher struct {
	queue   map[string]CompleteFuncInter
	rwmutex sync.RWMutex
}

func (f *Finisher) Clear() {
	f.queue = map[string]CompleteFuncInter{}
}

func (f *Finisher) Insert(finsher CompleteFuncInter) error {
	f.rwmutex.Lock()
	defer f.rwmutex.Unlock()
	if key, err := finsher.GetId(); err != nil {
		return err
	} else {
		f.queue[key] = finsher
	}
	return nil
}

func (f *Finisher) Remove(finsher CompleteFuncInter) error {
	f.rwmutex.Lock()
	defer f.rwmutex.Unlock()
	if key, err := finsher.GetId(); err != nil {
		return err
	} else {
		delete(f.queue, key)
	}
	return nil
}

func (f *Finisher) Execute() error {
	f.rwmutex.RLock()
	defer f.rwmutex.RUnlock()

	for key := range f.queue {
		if err := f.queue[key].Exit(); err != nil {
			return err
		}
	}
	return nil
}
