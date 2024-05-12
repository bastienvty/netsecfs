/*
 * JuiceFS, Copyright 2020 Juicedata, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package meta

import (
	"encoding/json"
	"fmt"
)

const (
	NoAtime = "noatime"
)

// Config for clients.
type Config struct {
	Strict     bool // update ctime
	MountPoint string
	AtimeMode  string
}

func DefaultConf() *Config {
	return &Config{Strict: true, AtimeMode: NoAtime}
}

type Format struct {
	Name      string
	UUID      string
	Storage   string
	BlockSize int
	Capacity  uint64 `json:",omitempty"`
}

func (f *Format) update(old *Format) error {
	var args []interface{}
	switch {
	case f.Name != old.Name:
		args = []interface{}{"name", old.Name, f.Name}
	case f.BlockSize != old.BlockSize:
		args = []interface{}{"block size", old.BlockSize, f.BlockSize}
	}
	if args == nil {
		f.UUID = old.UUID
	} else {
		return fmt.Errorf("cannot update volume %s from %v to %v", args...)
	}
	return nil
}

func (f *Format) String() string {
	t := *f
	s, _ := json.MarshalIndent(t, "", "  ")
	return string(s)
}
