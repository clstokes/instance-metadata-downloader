package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	var amazon bool
	var google bool

	flags := flag.NewFlagSet("instance-metadata-downloader", flag.ExitOnError)
	flags.SetOutput(os.Stdout)
	flags.Usage = usage

	flags.BoolVar(&amazon, "amazon", false, "amazon")
	flags.BoolVar(&google, "google", false, "google")

	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	outputPath := flags.Arg(0)
	if outputPath == "" {
		fmt.Printf("No output path specified.\n")
		usage()
	}

	if amazon {
		os.Exit(download(outputPath, nil))
	} else if google {
		os.Exit(download(outputPath, map[string]string{"Metadata-Flavor": "Google"}))
	} else {
		fmt.Printf("No provider specified.\n")
		usage()
	}

	os.Exit(0)
}

func download(outputPath string, headers map[string]string) int {
	host, url := "http://169.254.169.254", "/"
	data, err := recursiveGet(host, url, headers, make(map[string]string))

	if err != nil {
		fmt.Fprintf(os.Stderr, "err: [%v]\n", err.Error())
		return 1
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "err: [%v]\n", err.Error())
		return 1
	}

	writeMapToDisk(data, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: [%v]\n", err.Error())
		return 1
	}

	return 0
}

func recursiveGet(host string, url string, headers map[string]string, data map[string]string) (map[string]string, error) {
	fmt.Printf("Processing [%v]\n", url)

	resp, err := getRawResponse(host+url, headers)
	if err != nil {
		return data, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	if resp.StatusCode != 200 {
		// TODO: Tally these to return error at end of run.
		fmt.Printf("Received [%v] from [%v]. Saving anyway...\n", resp.StatusCode, url)
	}

	finalUrl := resp.Request.URL.Path

	if !strings.HasSuffix(finalUrl, "/") || resp.StatusCode != 200 {
		// process as normal file or error'd response
		fmt.Printf("Adding value from [%v]\n", finalUrl)
		data[finalUrl] = string(body)
		return data, nil
	} else {
		// descend into directory listing
		bodyString := string(body)
		data[finalUrl+"index.html"] = bodyString
		for _, val := range strings.Split(bodyString, "\n") {
			if val == "" {
				continue
			}
			data, err := recursiveGet(host, finalUrl+val, headers, data)
			if err != nil {
				return data, err
			}
		}
	}
	return data, nil
}

func getRawResponse(url string, headers map[string]string) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Duration(5 * time.Second), // TODO: make configurable?
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return client.Do(req)
}

func writeMapToDisk(data map[string]string, outputPath string) error {
	for key, val := range data {
		var err error
		filePath := outputPath + key

		dirSplit := strings.Split(filePath, "/")
		dir := strings.Join(dirSplit[0:len(dirSplit)-1], "/")

		fmt.Printf("Making directory: [%v]\n", dir)
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}

		fmt.Printf("Writing file: [%v]\n", filePath)
		err = ioutil.WriteFile(filePath, []byte(val), os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func usage() {
	helpText := `
usage: instance-metadata-downloader [options] path
Recursively downloads all available instance metadata to the given path.

Options:
  -amazon  Download Amazon instance metadata.
  -google  Download Google instance metadata.

`
	os.Stderr.WriteString(strings.TrimSpace(helpText) + "\n")
	os.Exit(1)
}
