package phantom

import "os/exec"

/*
Client provides interprocess communication with a custom phantomjs script.
*/
type Client struct {
	server *exec.Cmd
	port   int
}

/*
NewClient creates a new phantomjs subprocess and return a Client for querying
*/
func NewClient() *Client {
	return &Client{nil, 0}
}
