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

  flags := flag.NewFlagSet("instance-metadata-downloader", flag.ExitOnError)
  flags.SetOutput(os.Stdout)
  flags.Usage = usage

  flags.BoolVar(&amazon, "amazon", false, "amazon")

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
    os.Exit(RunAmazon(outputPath))
    return
  } else {
    fmt.Printf("No provider specified.\n")
    usage()
  }

  os.Exit(0)
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
  -amazon  Download amazon instance metadata.

`
  os.Stderr.WriteString(strings.TrimSpace(helpText) + "\n")
  os.Exit(1)
}
