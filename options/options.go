package options

import (
	"flag"
)

var cmdbBaseUrl = flag.String("cmdb_base_url", "http://cmdb-stackdev.inspurcloud.cn", "the base url of cmdb")

type Options struct {

	CmdbBaseUrl string

}

// NewOptions returns a new instance of `Options`.
func NewOptions() *Options {
	flag.Parse()
	return &Options{
		CmdbBaseUrl: *cmdbBaseUrl,
	}
}
