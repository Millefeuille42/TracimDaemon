# TracimDaemon

A master daemon for the TracimDaemon project

## Features

- Event based system
- "Local" Tracim event API
- "Local" Tracim API

## Usage

### Configuration

TracimDaemon will try to get a path to a config folder from the following selectors in order:

- First program argument
- User's config folder
- User's home folder + `.config/`

The selector fails if the element is not provided, by the user in the case of the `1st Argument` selector,
by the system in the case of the two others.

From now on the config folder will be referenced as `dir`.

TracimDaemon will the try to read the `dir/TracimDaemon`, if it does not exist, il will create it
along with a default config file and notification folder.


TracimDaemon is configured via a json configuration file, it is composed as follows:

```json
{
  "tracim": {
    "url": "http://localhost:8080/api",
    "username": "Me",
    "mail": "me@example.com",
    "password": "S3crƎtP4s$woRd"
  },
  "socket_path": "/path/to/sock"
}
```

- `tracim`: Information about the tracim server and user
  - `url`: URL of the tracim server, including the `api` route
  - `username`: Username of the tracim user, if required
  - `mail`: Email address of the tracim user, if required
  - `password`: Password of the tracim user
- `socket_path`: Path to the socket file

### Runtime

By itself, TracimDaemon does nothing besides logging incoming events.

To use it, you need to hook a plugin to it. For example, the [TracimPushNotification](https://github.com/Millefeuille42/TracimPushNotification) project.

When booting, plugins will notify TracimDaemon of their existence through the master socket and will be hooked to it.
Then, TracimDaemon will send them events as they come.

## Creating a plugin

See [TracimDaemonSDK](https://github.com/Millefeuille42/TracimDaemonSDK) for information on how to create a plugin.
