# TracimDaemon

A master daemon for the TracimDaemon project

## Features

- Event based system
- "Local" Tracim event API
- (WIP) "Local" Tracim API

## Usage

### Configuration

TracimDaemon is configured via environment variables.

- `TRACIM_DAEMON_TRACIM_URL`: URL of the Tracim server
- `TRACIM_DAEMON_TRACIM_MAIL`: If required by the Tracim server, the mail of the Tracim user
- `TRACIM_DAEMON_TRACIM_USERNAME`: If required by the Tracim server, the username of the Tracim user
- `TRACIM_DAEMON_TRACIM_PASSWORD`: The password of the Tracim user
- `TRACIM_DAEMON_SOCKET_PATH`: Path to the socket file

### Runtime

By itself, TracimDaemon does nothing besides logging incoming events.

To use it, you need to hook a plugin to it. For example, the [TracimPushNotification](https://github.com/Millefeuille42/TracimPushNotification) project.

When booting, plugins will notify TracimDaemon of their existence through the master socket and will be hooked to it.
Then, TracimDaemon will send them events as they come.

## Creating a plugin

See [TracimDaemonSDK](https://github.com/Millefeuille42/TracimPushNotification) for information on how to create a plugin.
