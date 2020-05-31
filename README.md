# openrvs-registry

## Features

This app enables the following features in OpenRVS:

- Fetching the latest version from GitHub
- Automatically updating a server list, based on active, healthy servers

## Listeners

It listens for HTTP requests on TCP port 8080, with the following endpoints:
- `/latest` returns the latest OpenRVS version from GitHub
- `/servers` returns a CSV list of game servers to OpenRVS clients
- `/servers/all` returns all servers, including unhealthy servers
- `/servers/debug` returns all servers with detailed health status information

It listens for OpenRVS beacons on UDP port 8080, for registration and health checking.

## Local Development

#### Initial Setup

1. Download and install the Go programming language for your OS here: https://golang.org/doc/install
1. Create a directory to contain all Go code, such as `%USERPROFILE%\go` (recommended)
1. Make sure the environment variable `GOPATH` is set to the above directory
1. Try to download openrvs-registry with `go get github.com/ijemafe/openrvs-registry` on the command line

#### Building and Running

Assuming a Windows development environment, there is a batch file to generate builds for both 64-bit Windows and 64-bit Linux at the same time:

```bash
> cd %GOPATH%\src\github.com\ijemafe\openrvs-registry\cmd\registry
> build.bat
```

The Windows build is `registry.exe`, and the Linux build is simply `registry`. Double click the `exe` to run the build locally. All log information is printed to `stdout` and displayed in the terminal window.

If you want to run locally without compiling a new build, you can:

```bash
> cd %GOPATH%\src\github.com\ijemafe\openrvs-registry\cmd\registry
> go run main.go
```

Tweak the code and repeat either set of steps above to iterate on changes.

## Deployments

There is an existing deployment of this software here:

- http://64.225.54.237:8080/servers
- http://64.225.54.237:8080/servers/all
- http://64.225.54.237:8080/servers/debug
- http://64.225.54.237:8080/latest
- udp://64.225.54.237:8080 (beacons)

If you need to stand up a new deployment:

1. Spin up a Linux or Windows server
1. Run the compiled softwware on the server, using some mechanism to keep the process running (`systemd` in Linux or `services.msc` in Windows)
1. Direct the output of the program into a log file. In `systemd`, for example, you can set the value of `StandardOutput` to `file:/full/path/to/registry.log` under `[Service]` in order to send all logs to that file. Logs will contain all information about healthchecks, status changes, saving to and loading from disk, and any errors which might occur.
1. Edit `OpenBeacon.uc` and change the host and port to the new registry server.
