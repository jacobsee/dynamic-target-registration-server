# Dynamic Target Registration Server

## What is it?

This is a small server that aims to make monitoring dynamic environments with Prometheus a bit easier. Environments can announce new systems that should be monitored using "beacons" - small apps that ping this server to announce their presence (and optionally pass some additional data, like security tokens). Similarly small apps called "handlers" watch this server for all beacon registrations of a specific type, and synchronize a directory of Prometheus target files to enable monitoring of those systems. If a system hosting a beacon "vanishes" without appropriately unregistering itself from this server first, Prometheus will consider that target to be `down`.

## How does it work?

This server exposes a simple HTTP interface with the following endpoints:

| Endpoint | Method | Description | Params |
| --- | --- | --- | --- |
| `/register` | `POST` | Registers the presence of a new beacon | `kind` - the type of system that is being registered</br>`key` - a unique string for identifying the beacon<br/>`data` - a stringified JSON object with any additional data necessary for monitoring |
| `/unregister` | `POST` | Cleanly unregisters a beacon to remove it from monitoring | `kind` - the type of system that is being unregistered<br/>`key` - the unique key that the beacon was registered with |
| `/list` | `GET` | Returns a JSON-encoded list of all active beacons of a certain type, and any arbitrary data that they were registered with | `kind` - the type of beacons to list |

**All** HTTP Requests must pass an `Authorization` header containing the string set in the `AUTH_TOKEN` environment variable, or they will be dropped.

## How is it configured?

This server takes the following environment variables:

| Variable Name | Description | Example |
| --- | --- | --- |
| `DB_FILE` | The location on disk where target information should be persisted so that it survives a server restart (note that this database is not human-readable) | `/some/dir/targets.db` |
| `AUTH_TOKEN` | A random string which clients will be expected to pass in an `Authorization` header to be able to communicate with this service | `12345abcde` |
