# Landlook — Interactive Landlock Profiler
![Landlook Demo](demo/demo.gif)

## How It Works
**Landlook** runs your application in a restricted **Landlock sandbox** and intercepts kernel audit events in real-time. When an action is blocked, it surfaces in an **interactive Terminal UI**, where you can instantly approve legitimate behaviors (file access, network calls, etc). By **iteratively restarting** the app with the updated profile and discovering hidden dependencies, you build a perfectly tailored **least-privilege security policy**.

## Requirements
 - Linux kernel `v6.15+` (for ABI v7 support)
 - `sudo` (for Netlink Audit only)

## Installation
Download from [Releases](https://github.com/cnaize/landlook/releases) or install via Go
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
   landlook - interactive landlock profiler

USAGE:
   landlook [global options] application [arguments]

GLOBAL OPTIONS:
   --log-level string                                           set zerolog level (default: error)
   --output string, -o string                                   output file (default: landlook.json)
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
