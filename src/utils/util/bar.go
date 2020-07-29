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
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	MAX_BAR_LENGTH = 30
	BAR_TYPE       = "#"
)

func NewBar(total int, prefix string, quiet bool) (*Bar, error) {
	bar := &Bar{}
	if err := bar.init(total, prefix, quiet); err != nil {
		return nil, err
	}
	return bar, nil
}

type Bar struct {
	eachOutNeedNum float32
	totalBar       int
	total          int
	outputNum      int
	finish         int
	rwmutex        sync.RWMutex
	quiet          bool
	haveShow       bool
	barId          string
	prefix         string
}

func (b *Bar) init(total int, prefix string, quiet bool) error {
	b.quiet = quiet
	if quiet {
		return nil
	}

	if total < 1 {
		return fmt.Errorf("Bar Init: Invalid arguments!")
	}
	b.total = total
	if total <= MAX_BAR_LENGTH {
		b.totalBar = total
		b.eachOutNeedNum = 1
	} else {
		b.eachOutNeedNum = float32(MAX_BAR_LENGTH) / float32(total)
		b.totalBar = MAX_BAR_LENGTH
	}
	b.prefix = prefix
	b.barId = b.prefix + strconv.FormatInt(time.Now().Unix(), 10) +
		strconv.Itoa(rand.Intn(100000)) + strconv.Itoa(b.total)
	return nil
}

func (b *Bar) show(isComplete bool) {
	b.haveShow = true
	showBar := fmt.Sprintf("%s |%s%s| %d/%d", b.prefix, strings.Repeat(BAR_TYPE, b.outputNum),
		strings.Repeat(" ", b.totalBar-b.outputNum), b.finish, b.total)
	if isComplete {
		fmt.Printf("\r%s\n", showBar)
	} else {
		fmt.Printf("\r%s", showBar)
	}
}

func (b *Bar) Finish(num int) {
	if b.quiet {
		return
	}
	b.rwmutex.Lock()
	defer b.rwmutex.Unlock()

	b.finish = num
	myOutputNum := int(float32(b.finish) * b.eachOutNeedNum)

	if myOutputNum > b.outputNum {
		if myOutputNum >= b.totalBar {
			b.outputNum = b.totalBar
		} else {
			b.outputNum = myOutputNum
		}
		b.show(false)
	}
}

func (b *Bar) Exit() error {
	if b.quiet {
		return nil
	}
	b.rwmutex.Lock()
	defer b.rwmutex.Unlock()
	if b.haveShow {
		b.show(true)
	}
	return nil
}

func (b *Bar) GetId() (string, error) {
	if b.barId == "" {
		return "", fmt.Errorf("the id of bar is empty!")
	}
	return b.barId, nil
}
