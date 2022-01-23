# openrvs-registry Developer Documentation

This repo uses [Go](https://golang.org/doc/install), a cross-platform, concurrent, compiled, garbage-collected, statically-typed programming language.

## Initial Setup

1. Download and install the Go programming language for your OS here: 
1. Clone this repo

## Navigating the Code

Currently, there are five files containing Go code:

1. `cmd/registry/main.go`: the primary code. starts the http and udp listeners,
	schedules disk checkpointing, and schedules healthchecks.
1. `csv.go`: contains code for converting CSV to Server objects and vice versa
1. `healthcheck.go`: contains logic for hiding unhealthy servers
1. `latest.go`: contains code for hitting the Github API
1. `types.go`: contains definitions and utility code unlikely to change

## Building the Code

Assuming a Windows development environment, there is a batch file to generate builds for both 64-bit Windows and 64-bit Linux at the same time:

```bash
> cd registry.git
> build.bat
```

The Windows build is `registry.exe`, and the Linux build is simply `registry`.

## Running the Code

Run `registry.exe` to run the build locally. All log information is printed to `stdout` and displayed in the terminal window:

```bash
> registry.exe
2020/05/30 23:35:27 openrvs-registry process started
2020/05/30 23:35:27 loading servers from file
2020/05/30 23:35:27 reading checkpoint file at checkpoint.csv
2020/05/30 23:35:27 there are now 48 registered servers (confirm over http)
2020/05/30 23:35:27 starting http listener
2020/05/30 23:35:27 starting udp listener
```

You can now hit the HTTP URLs in your browser at `http://localhost:8080/<path>`,
or send UDP beacons to `udp://localhost:8080` to test automatic registration.

If you want to run the app from a different working directory, you can:

```bash
> registry.exe -csvdir=C:\path\to\csv\files\\
```

The trailing slash must be included, and on Windows there must be two (since `\` is typically an escape character). On Linux, use forward slashes instead.

If you want to run locally without compiling a new build, you can:

```bash
> cd %GOPATH%\src\github.com\willroberts\openrvs-registry\cmd\registry
> go run main.go
```

Now you can tweak the code and repeat either set of steps above to iterate on changes.

## Deployments

There is an existing deployment at http://openrvs.org/servers

If you'd like to stand up a new deployment:

1. Build the code.
1. Spin up a Linux or Windows server.
1. Populate `seed.csv` on your server to choose some initial game servers.
1. Run the app on your server. Logs will be written to disk where the app is located.

## Pointing Clients at a Registry

Update the following values in `openrvs.ini`:

```
ServerURL openrvs.org:80 # or your IP and port
ServerListURL servers
```