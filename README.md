# openrvs-registry

Automatic server registration for OpenRVS.

For the client code, see [OpenRVS](https://github.com/OpenRVS-devs/OpenRVS).

## Features

This app enables the following features in OpenRVS:

- Automatically registering new OpenRVS servers as soon as they are started, with no human intervention
- Automatically hiding unhealthy servers after a configurable number of failed healthchecks
- Fetching the latest version from GitHub

## How It Works

When the app is first run, it looks for `seed.csv` as a source for the initial
server list data.

After populating the list in memory, the app begins sending healthchecks to each
known server on a regular interval. It uses these healthchecks to hide unhealthy
servers from the list (without fully removing them from memory; they continue
to receive healthchecks and may return if they become healthy again).

By default, healthchecks are sent every 30 seconds, and it takes 60 failed
checks to hide a server from the list (equivalent to 30 minutes downtime). A
single successful healthcheck will unhide the server.

Since OpenRVS v1.5, servers automatically send a REPORT beacon on startup
to a server running this app. When the app receives a beacon on its
UDP port, the information for that IP and port is updated. If the server is not
yet known, it is automatically added to the list.

## Listeners

There is a TCP listener for HTTP requests on port 8080, with the following endpoints:
- `/latest` returns the latest OpenRVS version from GitHub
- `/servers` returns a CSV list of game servers to OpenRVS clients
- `/servers/all` returns all servers, including unhealthy servers
- `/servers/debug` returns all servers with detailed health status information

There is also a UDP listener for OpenRVS beacons on port 8080, for registration and health checking.

## Deployments

There is an existing deployment of this software here:

- http://64.225.54.237:8080/servers
- http://64.225.54.237:8080/servers/all
- http://64.225.54.237:8080/servers/debug
- http://64.225.54.237:8080/latest
- udp://64.225.54.237:8080 (beacons)

If you'd like to stand up a new deployment:

1. Compile `openrvs-registry` with `build.bat` as described in the developer
docs.
1. Spin up a Linux or Windows server
1. Populate `seed.csv` based on the copy in this repo to choose the initial set
of servers. Put this file on the server alongside the compiled build.
1. Run the app on the server, using some mechanism to keep the process running
(either `systemd` in Linux or `services.msc` in Windows)
1. Direct the output of the program into a log file. In `systemd`, for example,
you can set the value of `StandardOutput` to `file:/full/path/to/registry.log`
under `[Service]` in order to send all logs to that file. Logs will contain all
information about healthchecks, status changes, saving to and loading from disk,
and any errors which might occur.
1. At this point, the server is running, has been seeded with the servers you
provided, and is now listening for registration beacons from OpenRVS servers.
1. To use the registry, edit `openrvs.ini` and change `ServerURL` to your
server's IP (followed by `:8080`) and change `ServerListURL` to `servers`.

## Developer Documentation

For developer docs, see [DOCS.md](DOCS.md).
