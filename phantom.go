package main

import "fmt"
import "net"
import "io"
import "os/exec"
import "strings"
import "strconv"

/*
ClientJob represents the data structure sent to phantomjs
*/
type ClientJob struct {
	URL string `json:"URL"`
}

/*
ClientSettings provides a clean interface for phantomjs CLI arguments
*/
type ClientSettings struct {
	LoadImages      bool
	IgnoreSSLErrors bool
}

/*
ToStringArgs returns the settings in a []string form for os/exec compatibility
*/
func (c *ClientSettings) ToStringArgs() []string {
	args := make([]string, 2, 2)

	if c.LoadImages {
		args[0] = "--load-images=true"
	} else {
		args[0] = "--load-images=false"
	}

	if c.IgnoreSSLErrors {
		args[1] = "--ignore-ssl-errors=true"
	} else {
		args[1] = "--ignore-ssl-errors=false"
	}

	return args
}

/*
DefaultSettings creates a new set of ClientSettings based on defaults
*/
func DefaultSettings() *ClientSettings {
	return &ClientSettings{false, true}
}

/*
Client provides interprocess communication with a custom phantomjs script.
*/
type Client struct {
	Server *exec.Cmd
	Port   uint64
	Out    io.ReadCloser
	ErrOut io.ReadCloser
}

/*
Close kills the underlying phantomjs process.
*/
func (c *Client) Close() {
	c.Server.Process.Kill()
}

/*
NewClient creates a new phantomjs subprocess and return a Client for querying.
*/
func NewClient(settings *ClientSettings) (*Client, error) {
	port, err := getAvailablePortNumber()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("phantomjs", "phantom.js", strconv.FormatUint(port, 10))
	cmd.Args = append(cmd.Args, settings.ToStringArgs()...)

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()

	if err != nil {
		return nil, err
	}

	return &Client{
		cmd,
		port,
		outPipe,
		errPipe,
	}, nil
}

func getAvailablePortNumber() (uint64, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	addr := ln.Addr().String()
	port := strings.Split(addr, ":")[strings.LastIndex(addr, ":")-1]
	return strconv.ParseUint(port, 10, 16)
}

func main() {
	client, err := NewClient(DefaultSettings())
	if err != nil {
		panic(err)
	}
	defer client.Close()

	fmt.Printf("PORT: %d\n", client.Port)
}
