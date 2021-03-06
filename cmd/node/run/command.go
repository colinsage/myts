package run

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"strings"

	"github.com/uber-go/zap"
	"log"
)

const logo = `
 8888888           .d888 888                   8888888b.  888888b.
   888            d88P"  888                   888  "Y88b 888  "88b
   888            888    888                   888    888 888  .88P
   888   88888b.  888888 888 888  888 888  888 888    888 8888888K.
   888   888 "88b 888    888 888  888  Y8bd8P' 888    888 888  "Y88b
   888   888  888 888    888 888  888   X88K   888    888 888    888
   888   888  888 888    888 Y88b 888 .d8""8b. 888  .d88P 888   d88P
 8888888 888  888 888    888  "Y88888 888  888 8888888P"  8888888P"

`

// Command represents the command executed by "influxd run".
type Command struct {
	Version   string
	Branch    string
	Commit    string
	BuildTime string

	closing chan struct{}
	Closed  chan struct{}

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Logger zap.Logger

	Server *Server
}

// NewCommand return a new instance of Command.
func NewCommand() *Command {
	return &Command{
		closing: make(chan struct{}),
		Closed:  make(chan struct{}),
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Logger:  zap.New(zap.NullEncoder()),
	}
}

// Run parses the config from args and runs the server.
func (cmd *Command) Run(args ...string) error {
	// Parse the command line flags.
	options, err := cmd.ParseFlags(args...)
	if err != nil {
		return err
	}

	// Print sweet InfluxDB logo.
	//fmt.Print(logo)
	cmd.Logger.Info(logo)

	// Set parallelism.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Mark start-up in log.
	cmd.Logger.Info(fmt.Sprintf("InfluxDB starting, version %s, branch %s, commit %s",
		cmd.Version, cmd.Branch, cmd.Commit))
	cmd.Logger.Info(fmt.Sprintf("Go version %s, GOMAXPROCS set to %d", runtime.Version(), runtime.GOMAXPROCS(0)))

	// Write the PID file.
	if err := cmd.writePIDFile(options.PIDFile); err != nil {
		return fmt.Errorf("write pid file: %s", err)
	}

	// Turn on block profiling to debug stuck databases
	runtime.SetBlockProfileRate(int(1 * time.Second))

	// Parse config
	config, err := cmd.ParseConfig(options.ConfigPath)
	if err != nil {
		return fmt.Errorf("parse config: %s", err)
	}


	// Propogate the top-level join options down to the meta config
	if config.Global.Join != "" {
		config.Meta.JoinPeers = strings.Split(config.Global.Join, ",")
	}

	// Command-line flags for -join and -hostname override the config
	// and env variable
	if options.Join != "" {
		config.Meta.JoinPeers = strings.Split(options.Join, ",")
	}

	if options.Hostname != "" {
		config.Global.Hostname = options.Hostname
	}

	// Propogate the top-level hostname down to dependendent configs
	config.Meta.RemoteHostname = config.Global.Hostname

	// Validate the configuration.
	if err := config.Validate(); err != nil {
		return fmt.Errorf("%s. To generate a valid configuration file run `influxd config > influxdb.generated.conf`", err)
	}

	// Create server from config and start it.
	buildInfo := &BuildInfo{
		Version: cmd.Version,
		Commit:  cmd.Commit,
		Branch:  cmd.Branch,
		Time:    cmd.BuildTime,
	}
	s, err := NewServer(config, buildInfo)
	if err != nil {
		return fmt.Errorf("create server: %s", err)
	}
	s.Logger = cmd.Logger


	if err := s.Open(); err != nil {
		return fmt.Errorf("open server: %s", err)
	}
	cmd.Server = s

	// Begin monitoring the server's error channel.
	go cmd.monitorServerErrors()

	return nil
}

// Close shuts down the server.
func (cmd *Command) Close() error {
	defer close(cmd.Closed)
	close(cmd.closing)
	if cmd.Server != nil {
		return cmd.Server.Close()
	}
	return nil
}

func (cmd *Command) monitorServerErrors() {
	logger := log.New(cmd.Stderr, "", log.LstdFlags)
	for {
		select {
		case err := <-cmd.Server.Err():
			logger.Println(err)
		case <-cmd.closing:
			return
		}
	}
}

// ParseFlags parses the command line flags from args and returns an options set.
func (cmd *Command) ParseFlags(args ...string) (Options, error) {
	var options Options
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringVar(&options.ConfigPath, "config", "", "")
	fs.StringVar(&options.PIDFile, "pidfile", "", "")
	fs.StringVar(&options.Join, "join", "", "")
	fs.StringVar(&options.Hostname, "hostname", "", "")
	fs.StringVar(&options.CPUProfile, "cpuprofile", "", "")
	fs.StringVar(&options.MemProfile, "memprofile", "", "")
	fs.Usage = func() { fmt.Fprintln(cmd.Stderr, usage) }
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	return options, nil
}

// writePIDFile writes the process ID to path.
func (cmd *Command) writePIDFile(path string) error {
	// Ignore if path is not set.
	if path == "" {
		return nil
	}

	// Ensure the required directory structure exists.
	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return fmt.Errorf("mkdir: %s", err)
	}

	// Retrieve the PID and write it.
	pid := strconv.Itoa(os.Getpid())
	if err := ioutil.WriteFile(path, []byte(pid), 0666); err != nil {
		return fmt.Errorf("write file: %s", err)
	}

	return nil
}

// ParseConfig parses the config at path.
// Returns a demo configuration if path is blank.
func (cmd *Command) ParseConfig(path string) (*Config, error) {
	// Use demo configuration if no config path is specified.
	if path == "" {
		cmd.Logger.Info("no configuration provided, using default settings")
		return NewDemoConfig()
	}

	cmd.Logger.Info(fmt.Sprintf("Using configuration at: %s", path))

	config := NewConfig()
	if err := config.FromTomlFile(path); err != nil {
		return nil, err
	}

	return config, nil
}

var usage = `usage: run [flags]

run starts the InfluxDB server. If this is the first time running the command
then a new cluster will be initialized unless the -join argument is used.

        -config <path>
                          Set the path to the configuration file.

        -join <host:port>
                          Joins the server to an existing cluster. Should be
                          the HTTP bind address of an existing meta server

        -hostname <name>
                          Override the hostname, the 'hostname' configuration
                          option will be overridden.

        -pidfile <path>
                          Write process ID to a file.

        -cpuprofile <path>
                          Write CPU profiling information to a file.

        -memprofile <path>
                          Write memory usage information to a file.
`

// Options represents the command line options that can be parsed.
type Options struct {
	ConfigPath string
	PIDFile    string
	Join       string
	Hostname   string
	CPUProfile string
	MemProfile string
}
