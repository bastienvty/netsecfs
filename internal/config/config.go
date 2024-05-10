package config

type FUSE struct {
	RootPath     string   `json:"root_path"`
	MountOptions []string `json:"mount_options,omitempty"`
	Name         string   `json:"display_name,omitempty"`

	EntryTimeout *int `json:"entry_timeout,omitempty"`
	AttrTimeout  *int `json:"attr_timeout,omitempty"`
}
