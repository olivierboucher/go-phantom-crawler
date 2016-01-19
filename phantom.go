package main

import "fmt"
import "net"
import "io"
import "bufio"

import "time"
import "os/exec"
import "strings"
import "bytes"
import "strconv"
import "net/http"
import "encoding/json"

import "github.com/satori/go.uuid"

/*
ClientJobResult represents the data structure retrieved from phantomjs
*/
type ClientJobResult struct {
	ID     string `json:"ID"`
	URL    string `json:"URL"`
	Result string `json:"result"`
}

/*
ClientJob represents the data structure sent to phantomjs
*/
type ClientJob struct {
	ID  string `json:"ID"`
	URL string `json:"URL"`
}

/*
NewJob creates a new crawling job data structure
*/
func NewJob(url string) *ClientJob {
	return &ClientJob{uuid.NewV4().String(), url}
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
	Server        *exec.Cmd
	Port          uint64
	CompletedJobs chan ClientJobResult
	StdOut        io.ReadCloser
	StdErr        io.ReadCloser
}

/*
Close kills the underlying phantomjs process.
*/
func (c *Client) Close() {
	c.Server.Process.Kill()
}

/*
QueueJob sends a job to the phantomjs process.
The result will be available through the Client's CompletedJobs channel.
*/
func (c *Client) QueueJob(job ClientJob) {
	go func(job ClientJob, client *Client) {
		data, _ := json.Marshal(job)
		buffer := bytes.NewBuffer(data)
		url := fmt.Sprintf("http://127.0.0.1:%d", client.Port)
		fmt.Println("URL: " + url)
		request, err := http.NewRequest("POST", url, buffer)

		if err != nil {
			panic("Could not create HTTP request")
		}

		request.Header.Set("Content-Type", "application/json")

		fmt.Println("About to send the request")
		response, err := http.DefaultClient.Do(request)
		fmt.Println("Got the response back")

		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			panic("Could not get a response from phantomjs")
		}

		defer response.Body.Close()
		fmt.Printf("Got response from phantomjs")

		decoder := json.NewDecoder(response.Body)
		var jobResult ClientJobResult
		err = decoder.Decode(&jobResult)

		if err != nil {
			panic("Could not unmarshal response body")
		}
		fmt.Printf("Completed job #%s.\nResult: %s\n", jobResult.ID, jobResult.Result)
		client.CompletedJobs <- jobResult

	}(job, c)
}

/*
NewClient creates a new phantomjs subprocess and return a Client for querying.
*/
func NewClient(settings *ClientSettings) (*Client, error) {
	/*port, err := getAvailablePortNumber()
	if err != nil {
		return nil, err
	}*/

	var port uint64
	port = 1337

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
		make(chan ClientJobResult),
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

	time.Sleep(3000 * time.Millisecond)

	stdOut := bufio.NewScanner(client.StdOut)
	stdErr := bufio.NewScanner(client.StdErr)

	go func() {
		fmt.Println("Starting to display Phantom std out")
		for stdOut.Scan() {
			line := stdOut.Text()
			fmt.Printf("Phantom out > %s\n", line)
		}
		fmt.Println("Stoping to display Phantom std out")
	}()

	go func() {
		fmt.Println("Starting to display Phantom std err")
		for stdErr.Scan() {
			line := stdErr.Text()
			fmt.Printf("Phantom err > %s\n", line)
		}
		fmt.Println("Stopping to display Phantom std err")
	}()

	go func() {
		err := client.Server.Wait()
		fmt.Printf("Phantomjs process finished: %v\n", err)
	}()

	job := NewJob("http://google.ca")
	client.QueueJob(*job)

	resp := <-client.CompletedJobs

	fmt.Printf("BODY: %s\n", resp.Result)
}
