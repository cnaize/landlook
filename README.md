# Landlook - Application inspection tool
![Landlook Demo](demo/demo.gif)

## How It Works
**Landlook** executes your application in a strict **Landlock sandbox** and intercepts kernel audit events in real-time. Any blocked action (files, network, etc) pops up in an **interactive Terminal UI**, where you can instantly analyze the application's hidden requirements. By **iteratively restarting** the app and approving legitimate behaviors, you discover exactly what permissions the process needs to function, making it the ultimate tool for **security profiling** and **debugging** modern **Linux applications**.

## Requirements
 - Linux kernel `v6.15+` (for ABI v7 support)
 - `sudo` (for Netlink Audit only)

## Installation
```bash
go install github.com/cnaize/landlook/cmd/landlook@latest
```

## Example Usage
```bash
sudo landlook -- ls -la /tmp
```

## Command-line options
```text
NAME:
   landlook - application inspection tool

USAGE:
   landlook [global options] application [arguments]

GLOBAL OPTIONS:
   --log-level string                                           set zerolog level (default: error)
   --ro string [ --ro string ]                                  allow read/exec path (default: deny all)
   --rw string [ --rw string ]                                  allow read/exec/write path (default: deny all)
   --tcp-listen uint, -l uint [ --tcp-listen uint, -l uint ]    allow listen tcp port (default: deny all)
   --tcp-connect uint, -c uint [ --tcp-connect uint, -c uint ]  allow connect tcp port (default: deny all)
   --sockets                                                    allow open abstract sockets (default: deny)
   --signals                                                    allow send signals (default: deny)
   --env string, -e string [ --env string, -e string ]          add environment variable (default: empty list)
   --add-self                                                   add application itself to --ro (default: true)
   --add-deps                                                   add application dependencies to --ro (default: true)
   --help, -h                                                   show help
```

## Features
 - [x] Linux amd64 support
 - [ ] Linux arm64 support
 - [ ] Export profile to JSON/YAML
