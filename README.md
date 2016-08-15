# instance-metadata-downloader

`instance-metadata-downloader` recursively downloads instance metadata for offline
use.

## Usage

```
$ instance-metadata-downloader
usage: instance-metadata-downloader [options] path
Recursively downloads all available instance metadata to the given path.

Options:
  -amazon  Download Amazon instance metadata.
  -google  Download Google instance metadata.

$ instance-metadata-downloader -amazon amazon-data
Processing [/]
Processing [/1.0]
Processing [/1.0/meta-data]
Processing [/1.0/meta-data/ami-id]
Adding value from [/1.0/meta-data/ami-id]
Processing [/1.0/meta-data/ami-launch-index]
Adding value from [/1.0/meta-data/ami-launch-index]
...

$ tree amazon-data
amazon-data/
├── 1.0
│   ├── index.html
│   ├── meta-data
│   │   ├── ami-id
│   │   ├── ami-launch-index
...
```

## FAQ

### What providers are supported?

```
amazon
google
```

### Why not just use `wget --recursive`?

Most instance metadata services are strictly text-based and do not provide
navigable HTML links in their output. As such, the output must be parsed
in order to be navigated and downloaded.
