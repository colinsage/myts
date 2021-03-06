package run

import (
	"github.com/BurntSushi/toml"
	"os/user"
	"os"
	"fmt"
	"path/filepath"
	"io/ioutil"
	"bytes"
	"regexp"
	"strings"
	"log"
	"github.com/influxdata/influxdb/tsdb"
	"github.com/influxdata/influxdb/services/precreator"
	"github.com/influxdata/influxdb/monitor"

	"github.com/colinsage/myts/services/meta"
	"github.com/colinsage/myts/services/data"

	"github.com/colinsage/myts/services/hh"
	"github.com/colinsage/myts/services/retention"
	"github.com/colinsage/myts/services/httpd"
)

// Config represents the configuration format for the influxd binary.
type Config struct {
	// Hostname is the hostname portion to use when registering local
	// addresses.  This hostname must be resolvable from other nodes.

	Global  *Global `toml:"global"`

	Meta    *meta.Config `toml:"meta"`
	Data    tsdb.Config  `toml:"data"`
	DataExt *data.Config `toml:"data-ext"`

	HH hh.Config  `toml:"hinted-handoff"`

	Retention   retention.Config   `toml:"retention"`
	Precreator  precreator.Config  `toml:"shard-precreation"`
	Monitor        monitor.Config    `toml:"monitor"`



	HTTPD          httpd.Config      `toml:"http"`


	// BindAddress is the address that all TCP services use (Raft, Snapshot, DataExt, etc.)
	//BindAddress string `toml:"bind-address"`
	BindAddress  string  `toml:"-"`


}

type Global struct {
	Hostname string `toml:"hostname"`
	Join string `toml:"join"`

	MetaEnabled bool  `toml:"meta-enable"`
	DataEnabled bool  `toml:"data-enable"`

	LogPath string  `toml:"log-path"`
}

// NewConfig returns an instance of Config with reasonable defaults.
func NewConfig() *Config {
	c := &Config{}
	c.Global = &Global{}
	c.Meta = meta.NewConfig()
	c.Data = tsdb.NewConfig()
	c.DataExt = data.NewConfig()
	c.HH = hh.NewConfig()

	c.Monitor = monitor.NewConfig()
	c.HTTPD = httpd.NewConfig()
	c.Precreator = precreator.NewConfig()

	c.Retention = retention.NewConfig()

	return c
}

// Validate returns an error if the config is invalid.
func (c *Config) Validate() error {
	if c.Global.MetaEnabled {
		if err := c.Meta.Validate(); err != nil {
			return err
		}
	}

	if c.Global.DataEnabled {
		if err := c.Data.Validate(); err != nil {
			return err
		}
	}


	return nil
}

// NewDemoConfig returns the config that runs when no config is specified.
func NewDemoConfig() (*Config, error) {
	c := NewConfig()

	var homeDir string
	// By default, store meta and data files in current users home directory
	u, err := user.Current()
	if err == nil {
		homeDir = u.HomeDir
	} else if os.Getenv("HOME") != "" {
		homeDir = os.Getenv("HOME")
	} else {
		return nil, fmt.Errorf("failed to determine current user for storage")
	}

	c.Meta.Dir = filepath.Join(homeDir, "var/influxdb/meta")

	return c, nil
}

func (c *Config) FromTomlFile(fpath string) error {
	bs, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}
	bs = trimBOM(bs)
	return c.FromToml(string(bs))
}

// trimBOM trims the Byte-Order-Marks from the beginning of the file.
// This is for Windows compatability only.
// See https://github.com/influxdata/telegraf/issues/1378.
func trimBOM(f []byte) []byte {
	return bytes.TrimPrefix(f, []byte("\xef\xbb\xbf"))
}
// FromToml loads the config from TOML.
func (c *Config) FromToml(input string) error {
	// Replace deprecated [cluster] with [coordinator]
	re := regexp.MustCompile(`(?m)^\s*\[cluster\]`)
	input = re.ReplaceAllStringFunc(input, func(in string) string {
		in = strings.TrimSpace(in)
		out := "[coordinator]"
		log.Printf("deprecated config option %s replaced with %s; %s will not be supported in a future release\n", in, out, in)
		return out
	})

	_, err := toml.Decode(input, c)
	return err
}