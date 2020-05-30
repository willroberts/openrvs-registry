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

## Deployments

So far I have only deployed this locally. Eventually we can put this on a web
server behind Nginx.