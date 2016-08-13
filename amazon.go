package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func RunAmazon(outputPath string) int {
	host, url := "http://169.254.169.254", "/"
	data, err := recursiveGet(host, url, make(map[string]string))

	if err != nil {
		fmt.Fprintf(os.Stderr, "err: [%s]\n", err.Error())
		return 1
	}

	fmt.Printf("Final data: [%s]\n", data)

	if err != nil {
		fmt.Fprintf(os.Stderr, "err: [%s]\n", err.Error())
		return 1
	}

	WriteMapToDisk(data, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: [%s]\n", err.Error())
		return 1
	}

	return 0
}

func recursiveGet(host string, url string, data map[string]string) (map[string]string, error) {
	fmt.Printf("Processing [%s]\n", url)

	resp, err := GetRawResponse(host+url, nil)
	if err != nil {
		return data, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	finalUrl := resp.Request.URL.Path

	if !strings.HasSuffix(finalUrl, "/") {
		// process as normal file
		fmt.Printf("Adding value from [%s]\n", finalUrl)
		data[finalUrl] = string(body)
		return data, nil
	} else {
		// descend into directory listing
		bodyString := string(body)
		data[finalUrl+"/index.html"] = bodyString
		for _, val := range strings.Split(bodyString, "\n") {
			if val == "" {
				continue
			}
			data, err := recursiveGet(host, finalUrl+val, data)
			if err != nil {
				return data, err
			}
		}
	}
	return data, nil
}
