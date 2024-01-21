# OpenRVS Registry Developer Docs

This repo uses Go, a cross-platform, concurrent, compiled, garbage-collected, statically-typed programming language.

## Initial Setup

1. [Download and install the Go programming language for your OS](https://golang.org/doc/install)
1. Clone this repo

## Building the Code

**On Mac or Linux:**
```bash
# Output will be in bin/registry.
make build
```

**On Windows:**
```bat
cd openrvs-registry.git\cmd\registry
REM Output will be in cmd/registry/registry.exe.
build.bat
```

## Running the Code

Run `registry[.exe]` to run the build locally. All log information is printed to `stdout` and displayed in the terminal window:

```bash
> registry.exe
2020/05/30 23:35:27 openrvs-registry process started
2020/05/30 23:35:27 loading servers from file
2020/05/30 23:35:27 reading checkpoint file at checkpoint.csv
2020/05/30 23:35:27 there are now 48 registered servers (confirm over http)
2020/05/30 23:35:27 starting http listener
2020/05/30 23:35:27 starting udp listener
```

You can now hit the HTTP URLs in your browser (e.g. `http://localhost:8080/servers`),
or send UDP beacons to `udp://localhost:8080` to test automatic registration.

If you want to run the app from a different working directory, you can use:
```bat
registry.exe -csvdir=C:\path\to\csv\files\\
```

The trailing slash must be included, and on Windows there must be two (since `\` is typically an escape character). On Linux, use forward slashes instead.

If you want to run locally without compiling a new build, you can use:
```bat
cd openrvs-registry.git\cmd\registry
go run main.go
```

## Deployments

There is an existing deployment at http://openrvs.org/servers

If you'd like to stand up a new deployment:

1. Build the code and copy it to your Linux or Windows server.
1. Populate `seed.csv` to choose some initial game servers.
1. Run the app.

## Pointing OpenRVS Clients at a Registry server

Update the following values in `openrvs.ini`:
```
ServerURL openrvs.org:80 # or your IP and port
ServerListURL servers
```
