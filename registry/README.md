# OpenRVS Registry v2

## Design

- UDP server on port 8080 for automatic registration
- Models for GameServerInfo and GameServerHealth
- Registry class has methods to load/save from/to disk, add/remove server entries
  - RegistryConfig class within it
- HTTP server on port 8080 for serving the userlist, latest version, and manual registration
- CSV serializer class to handle the checkpoints
- cmd/v2/main.go provides entrypoint & initial config

## To Do

- Design and commit interfaces
- Migrate one interface at a time to v2
