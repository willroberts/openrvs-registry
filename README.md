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

## Adding new servers

As of OpenRVS 1.5, new servers will automatically register themselves with the registry.

Servers can also be manually added over HTTP:

```
$ curl -X POST https://openrvs.org/servers/add -d "host:port"
```

## Listeners

There is a TCP listener for HTTP requests on port 8080, with the following endpoints:
- `/latest` returns the latest OpenRVS version from GitHub
- `/servers` returns a CSV list of game servers to OpenRVS clients
- `/servers/all` returns all servers, including unhealthy servers
- `/servers/debug` returns all servers with detailed health status information

There is also a UDP listener for OpenRVS beacons on port 8080, for registration and health checking.

## Developer Documentation

For developer docs, see [DOCS.md](DOCS.md).
