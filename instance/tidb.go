// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package instance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pingcap/tiup/pkg/repository/v0manifest"
	"github.com/pingcap/tiup/pkg/utils"
)

// TiDBInstance represent a running tidb-server
type TiDBInstance struct {
	instance
	Process
	enableBinlog bool
}

// NewTiDBInstance return a TiDBInstance
func NewTiDBInstance(binPath string, dir, host, configPath string, id int, enableBinlog bool) *TiDBInstance {
	return &TiDBInstance{
		instance: instance{
			BinPath:    binPath,
			ID:         id,
			Dir:        dir,
			Host:       host,
			Port:       utils.MustGetFreePort(host, 4000),
			StatusPort: utils.MustGetFreePort("0.0.0.0", 10080),
			ConfigPath: configPath,
		},
		enableBinlog: enableBinlog,
	}
}

// Start calls set inst.cmd and Start
func (inst *TiDBInstance) Start(ctx context.Context, version v0manifest.Version) error {
	if err := os.MkdirAll(inst.Dir, 0755); err != nil {
		return err
	}
	args := []string{
		"-P", strconv.Itoa(inst.Port),
		"--store=mocktikv",
		fmt.Sprintf("--host=%s", inst.Host),
		fmt.Sprintf("--status=%d", inst.StatusPort),
		fmt.Sprintf("--log-file=%s", filepath.Join(inst.Dir, "tidb.log")),
	}
	if inst.ConfigPath != "" {
		args = append(args, fmt.Sprintf("--config=%s", inst.ConfigPath))
	}
	if inst.enableBinlog {
		args = append(args, "--enable-binlog=true")
	}

	var err error
	if inst.Process, err = NewComponentProcess(ctx, inst.Dir, inst.BinPath, "tidb", version, args...); err != nil {
		return err
	}
	logIfErr(inst.Process.SetOutputFile(inst.LogFile()))

	return inst.Process.Start()
}

// Component return the component name.
func (inst *TiDBInstance) Component() string {
	return "tidb"
}

// LogFile return the log file name.
func (inst *TiDBInstance) LogFile() string {
	return filepath.Join(inst.Dir, "tidb.log")
}

// Addr return the listen address of TiDB
func (inst *TiDBInstance) Addr() string {
	return fmt.Sprintf("%s:%d", advertiseHost(inst.Host), inst.Port)
}