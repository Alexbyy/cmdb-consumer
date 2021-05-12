package manager

import (
	"inspur.com/cmdb-consumer/cmdb"
	"inspur.com/cmdb-consumer/options"
)

type Manager struct {
	Agent   *cmdb.Client
	Options *options.Options
}


func NewManager(a *cmdb.Client, opts *options.Options) (*Manager, error) {
	return &Manager{
		Agent:   a,
		Options: opts,
	}, nil
}

