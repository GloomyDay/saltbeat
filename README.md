# Saltbeat

Welcome to Saltbeat.
## Getting Started with Saltbeat

### Requirements

* [Golang](https://golang.org/dl/) 1.18
### Build

To build the binary for Saltbeat run the command below. This will generate a binary
in the same directory with the name saltbeat.

```
./build.sh
```

### Run

To run Saltbeat with debugging output enabled, run:

```
./saltbeat -c saltbeat.yml -e -d "*"
```

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).
