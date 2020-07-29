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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

import (
	"github.com/alecthomas/kingpin"
)

import (
	"argparser"
	"bceconf"
	"utils/util"
)

type bceConfig struct {
	configPath string
}

func (b *bceConfig) bceReloadConfigPath(c *kingpin.ParseContext) error {
	bceconf.ReloadConfAction(b.configPath)
	return nil
}

func (b *bceConfig) configInteractive(context *kingpin.ParseContext) error {
	bceconf.ConfigInteractive(b.configPath)
	os.Exit(0)
	return nil
}

func showVersion(c *kingpin.ParseContext) error {
	fmt.Printf("bcecmd v%s from http://bce.baidu.com\n", bceconf.BCE_VERSION)
	os.Exit(0)
	return nil
}

func setDebug(c *kingpin.ParseContext) error {
	bceconf.SetDebug()
	return nil
}

func buildArgumentParser(bcecmd *kingpin.Application) {
	b := &bceConfig{}

	bcecmd.Flag(
		"configure",
		"configure AK SK Region and Domain for bcecmd and will be written to "+
			"CONF_PATH(the user's home director by default").
		Short('c').Action(b.configInteractive).Default("").StringVar(&b.configPath)

	bcecmd.Flag(
		"debug",
		"cli debug").
		Short('d').Action(setDebug).Bool()

	bcecmd.Flag(
		"conf-path",
		"config path").
		StringVar(&b.configPath)

	bcecmd.Flag(
		"version",
		"show program's version number and exit").
		Short('v').Action(showVersion).Bool()

	bos := bcecmd.Command(
		"bos",
		"bos service").
		PreAction(b.bceReloadConfigPath)

	argparser.BuildBosParser(bos)

	bosApi := bcecmd.Command(
		"bosapi",
		"BOS API command.").
		PreAction(b.bceReloadConfigPath)

	argparser.BuildBosApi(bosApi)

	bosProbeCmd := bcecmd.Command(
		"bosprobe",
		"bos probe").
		PreAction(b.bceReloadConfigPath)

	argparser.BuildBosProbeParser(bosProbeCmd)
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Error: %v!\n", err)
			os.Exit(1)
		}
	}()

	// handling interrupt (ctrl+c)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		if util.GFinisher != nil {
			err := util.GFinisher.Execute()
			if err != nil {
				fmt.Printf("Error: %v!\n", err)
			}
		}
		os.Exit(1)
	}()

	bcecmd := kingpin.New("bcecmd", "BCE Command Line Interface")
	buildArgumentParser(bcecmd)

	_, err, context := bcecmd.Parse(os.Args[1:])
	if err != nil {
		bcecmd.WriteUsage(context, err)
	}

	bceconf.DestructConfFolder()
}
