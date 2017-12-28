package meta

import (
	"time"
	"github.com/influxdata/influxdb/toml"
	"runtime"
	"path"
	"io/ioutil"
	"os"
	"testing"
	"log"
)

func Test_CreateDatabase(t *testing.T) {
	t.Parallel()

	config := newConfig()
	client := newClient(config)
	defer os.RemoveAll(config.Dir)

	defer client.Close()
	dbname := "db3"
	if _, err := client.CreateDatabase(dbname); err != nil {
		t.Fatal(err)
	}

	db := client.Database(dbname)
	if db == nil {
		t.Fatal("no database got.")
	} else if db.Name != dbname {
		t.Fatalf("db name wrong: %s", db.Name)
	}

	// Make sure a default retention policy was created.
	_, err := client.RetentionPolicy(dbname, "default")
	if err != nil {
		t.Fatal(err)
	} else if db.DefaultRetentionPolicy != "default" {
		t.Fatalf("rp name wrong: %s", db.DefaultRetentionPolicy)
	}
}

func newClient(config *Config) *Client {
	c := NewClient()

	c.SetMetaServers([]string{"127.0.0.1:7091","127.0.0.1:8091","127.0.0.1:9091",})
	if err := c.Open(); err != nil {
		panic(err)
	}
	log.Println("cluster open done")
	return c
}

func newConfig() *Config {
	cfg := NewConfig()
	cfg.BindAddress = "127.0.0.1:7088"
	cfg.HTTPBindAddress = "127.0.0.1:8091"
	cfg.Dir = testTempDir(2)
	cfg.LeaseDuration = toml.Duration(1 * time.Second)
	return cfg
}

func testTempDir(skip int) string {
	// Get name of the calling function.
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		panic("failed to get name of test function")
	}
	_, prefix := path.Split(runtime.FuncForPC(pc).Name())
	// Make a temp dir prefixed with calling function's name.
	dir, err := ioutil.TempDir(os.TempDir(), prefix)
	if err != nil {
		panic(err)
	}
	return dir
}