# openrvs-registry

## Features

This app enables the following features in OpenRVS:

- Fetching the latest version from GitHub
- Automatically updating a server list, based on active, healthy servers

## Listeners

It listens for HTTP requests on TCP port 8080, with the following endpoints:
- `/servers` returns a CSV list of game servers to OpenRVS clients
- `/latest` returns the latest OpenRVS version from GitHub
- `/servers/unhealthy` returns servers currently hidden from the set (coming soon)

It listens for OpenRVS beacons on UDP port 8080, for registration and health checking.

## To Do

- Finish health checking

## Deployments

There is an existing deployment of this software here:

- http://64.225.54.237:8080/servers
- http://64.225.54.237:8080/latest
- udp://64.225.54.237:8080 (beacons)

## Management

If needed, we can have an extra HTTP listener bound to `127.0.0.1:<some_port>`,
attach administrative handlers to that listener (e.g. `/delete_server`), and
then use a web server like Nginx to handle authorization over the public
Internet.

Currently, the only administrative action is manually changing the server list,
which should not be necessary under normal circumstances. To do this:

```bash
$ systemctl stop registry # stop the app
$ vim checkpoint.csv # edit checkpoint.csv and/or seed.csv as desired
$ systemctl start registry # start the app
```