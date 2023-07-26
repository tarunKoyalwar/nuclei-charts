# nuclei-charts

Nuclei Charts is a simple utility to generate charts for testing purposes from nuclei stats.json file

### Usage

- Checkout nuclei branch `feat-perf-testing` 
- build nuclei with `stats` tag
```console
$ go build -tags stats .
```
- Run nuclei with 1 target (only 1 target is supported for now) with desired templates and flags
```console
$ ./nuclei -u https://scanme.sh -stats -c 2000 -rl 5000 
```
- ^ will generate a `stats.json` file in the current directory


### Install nuclei-charts

```console
$ go install -v github.com/tarunKoyalwar/nuclei-charts@latest
```

or build from source

## Generate Charts

### To start a webserver on port 8081 with charts run

```console
$ nuclei-charts -input stats.json
```

### To generate a HTML file with charts run

```console
$ nuclei-charts -input stats.json -output output.html
```
