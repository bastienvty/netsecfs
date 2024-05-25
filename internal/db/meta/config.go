package meta

import (
	"encoding/json"
	"fmt"
)

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
