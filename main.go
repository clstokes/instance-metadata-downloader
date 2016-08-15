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
    fmt.Printf("%s\n", err)
    return
  }

  outputPath := flags.Arg(0)
  if outputPath == "" {
    fmt.Printf("No output path specified.\n")
    usage()
  }

  if amazon {
    os.Exit(Download(outputPath, nil))
  } else if google {
    os.Exit(Download(outputPath, map[string]string{"Metadata-Flavor": "Google"}))
  } else {
    fmt.Printf("No provider specified.\n")
    usage()
  }

  os.Exit(0)
}

func Download(outputPath string, headers map[string]string) int {
  host, url := "http://169.254.169.254", "/"
  data, err := recursiveGet(host, url, headers, make(map[string]string))

  if err != nil {
    fmt.Fprintf(os.Stderr, "err: [%s]\n", err.Error())
    return 1
  }

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

func recursiveGet(host string, url string, headers map[string]string, data map[string]string) (map[string]string, error) {
  fmt.Printf("Processing [%s]\n", url)

  resp, err := GetRawResponse(host+url, headers)
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
      data, err := recursiveGet(host, finalUrl+val, headers, data)
      if err != nil {
        return data, err
      }
    }
  }
  return data, nil
}

func GetRawResponse(url string, headers map[string]string) (*http.Response, error) {
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

func WriteMapToDisk(data map[string]string, outputPath string) error {
  for key, val := range data {
    var err error
    filePath := outputPath + key

    dirSplit := strings.Split(filePath, "/")
    dir := strings.Join(dirSplit[0:len(dirSplit)-1], "/")

    fmt.Printf("Making directory: [%s]\n", dir)
    err = os.MkdirAll(dir, os.ModePerm)
    if err != nil {
      return err
    }

    fmt.Printf("Writing file: [%s]\n", filePath)
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
