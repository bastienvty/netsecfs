# netsecfs

A shared encrypted network file system.

## Platforms

The application has been tested on Ubuntu 22.04 LTS and macOS Sonoma 14.5 on M2.
It should work fine both on ARM and x86 architectures. If you encounter any issues, please contact me.

## Prerequisites

The application requires the following packages to be installed on the system.
The version of go should be at least is 1.22.2.

Follow the official guidelines to install go on your system: [golang.org](https://golang.org/doc/install).

### Ubuntu

Install the following packages:

```bash
$ sudo apt install fuse
```

### macOS

Install [macFUSE](https://osxfuse.github.io/). The latest version (4.7.2) should work just fine.

## Compile

To compile the application, run the following command:

```bash
$ git clone https://github.com/bastienvty/netsecfs
$ cd netsecfs
$ go build .
```

## Usage

Example of how to initialise and mount a file system.

```bash
$ ./netsecfs init --storage data.db --meta meta.db myfs
$ ./netsecfs --meta meta.db /tmp/nsfs
```

We can now interact with the CLI of the application.

```bash
netsecfs> signup test admin
netsecfs> mount
```

The file system is now mounted at `/tmp/nsfs` as user `test`.

To get a list of all available commands, type `help`.

## Warning

This application is a proof of concept and should not be used in production.