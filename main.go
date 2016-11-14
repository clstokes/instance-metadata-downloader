package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
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
	data := make(map[string]string)
	err := recursiveGet(host, url, headers, &data)

	if err != nil {
		fmt.Fprintf(os.Stderr, "err: [%v]\n", err.Error())
		return 1
	}

	err = writeMapToDisk(&data, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: [%v]\n", err.Error())
		return 1
	}

	return 0
}

func recursiveGet(host string, url string, headers map[string]string, data *map[string]string) error {
	fmt.Printf("debug: Processing [%v]\n", url)

	resp, err := getRawResponse(host+url, headers)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		// TODO: Tally these to return error at end of run.
		fmt.Printf("warn: Received [%v] from [%v]. Saving anyway...\n", resp.StatusCode, url)
	}

	finalUrl := resp.Request.URL.Path

	if !strings.HasSuffix(finalUrl, "/") || resp.StatusCode != 200 {
		// process as normal file or error'd response
		fmt.Printf("debug: Adding value from [%v]\n", finalUrl)
		(*data)[finalUrl] = string(body)
		return nil
	} else {
		// descend into directory listing
		bodyString := string(body)
		(*data)[finalUrl+"index.html"] = bodyString
		for _, val := range strings.Split(bodyString, "\n") {
			if val == "" {
				continue
			}
			err := recursiveGet(host, finalUrl+val, headers, data)
			if err != nil {
				return err
			}
		}
	}
	return nil
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

func writeMapToDisk(data *map[string]string, outputPath string) error {
	// create slice from keys for sorting
	var keys []string
	for key, _ := range *data {
		keys = append(keys, key)
	}

	// sort keys based on depth of path
	sort.Sort(SortByPaths(keys))

	for _, key := range keys {
		val := (*data)[key]
		var err error
		filePath := outputPath + key

		dirSplit := strings.Split(filePath, "/")
		dir := strings.Join(dirSplit[0:len(dirSplit)-1], "/")

		fmt.Printf("debug: Making directory: [%v]\n", dir)
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			if strings.Contains(err.Error(), "not a directory") {
				fmt.Fprintf(os.Stderr, "warn: [%v]. Continuing...\n", err.Error())
			} else {
				return err
			}
		}

		fmt.Printf("debug: Writing file: [%v]\n", filePath)
		err = ioutil.WriteFile(filePath, []byte(val), os.ModePerm)
		if err != nil {
			if strings.Contains(err.Error(), "is a directory") {
				fmt.Fprintf(os.Stderr, "warn: [%v]. Continuing...\n", err.Error())
			} else {
				return err
			}
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

type SortByPaths []string

func (s SortByPaths) Len() int {
	return len(s)
}
func (s SortByPaths) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortByPaths) Less(i, j int) bool {
	return strings.Count(s[i], "/") > strings.Count(s[j], "/")
}
