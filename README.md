<!--
title: go.d.plugin
description: "go.d.plugin is an external plugin for Netdata, responsible for running individual data collectors written in Go."
custom_edit_url: https://github.com/netdata/go.d.plugin/edit/master/README.md
sidebar_label: "go.d.plugin"
learn_status: "Published"
learn_topic_type: "Tasks"
learn_rel_path: "Developers/External plugins/go.d.plugin"
sidebar_position: 1
-->

# go.d.plugin

`go.d.plugin` is a [Netdata](https://github.com/netdata/netdata) external plugin. It is an **orchestrator** for data
collection modules written in `go`.

1. It runs as an independent process (`ps fax` shows it).
2. It is started and stopped automatically by Netdata.
3. It communicates with Netdata via a unidirectional pipe (sending data to the Netdata daemon).
4. Supports any number of data collection [modules](https://github.com/netdata/go.d.plugin/tree/master/modules).
5. Allows each [module](https://github.com/netdata/go.d.plugin/tree/master/modules) to have any number of data
   collection jobs.

## Bug reports, feature requests, and questions

Are welcome! We are using [netdata/netdata](https://github.com/netdata/netdata/) repository for bugs, feature requests,
and questions.

- [GitHub Issues](https://github.com/netdata/netdata/issues/new/choose): report bugs or open a new feature request.
- [GitHub Discussions](https://github.com/netdata/netdata/discussions): ask a question or suggest a new idea.

## Install

Go.d.plugin is shipped with Netdata.

### Required Linux capabilities

All capabilities are set automatically during Netdata installation using
the [official installation method](https://github.com/netdata/netdata/blob/master/packaging/installer/README.md#install-on-linux-with-one-line-installer).
No further action required. If you have used a different installation method and need to set the capabilities manually,
see the appropriate collector readme.

| Capability          |                                       Required by                                        |
|:--------------------|:----------------------------------------------------------------------------------------:|
| CAP_NET_RAW         |      [Ping](https://github.com/netdata/go.d.plugin/tree/master/modules/ping#readme)      |
| CAP_NET_ADMIN       | [Wireguard](https://github.com/netdata/go.d.plugin/tree/master/modules/wireguard#readme) |
| CAP_DAC_READ_SEARCH | [Filecheck](https://github.com/netdata/go.d.plugin/tree/master/modules/filecheck#readme) |

## Available modules

| Name                                                                                                |           Monitors            |
|:----------------------------------------------------------------------------------------------------|:-----------------------------:|
| docker_network | Docker Networking |


## Configuration

Edit the `go.d.conf` configuration file using `edit-config` from the
Netdata [config directory](https://github.com/netdata/netdata/blob/master/docs/configure/nodes.md), which is typically
at `/etc/netdata`.

```bash
cd /etc/netdata # Replace this path with your Netdata config directory
sudo ./edit-config go.d.conf
```

Configurations are written in [YAML](http://yaml.org/).

- [plugin configuration](https://github.com/netdata/go.d.plugin/blob/master/config/go.d.conf)
- [specific module configuration](https://github.com/netdata/go.d.plugin/tree/master/config/go.d)

### Enable a collector

To enable a collector you should edit `go.d.conf` to uncomment the collector in question and change it from `no`
to `yes`.

For example, to enable the `example` plugin you would need to update `go.d.conf` from something like:

```yaml
modules:
#  example: no 
```

to

```yaml
modules:
  example: yes
```

Then [restart netdata](https://github.com/netdata/netdata/blob/master/docs/configure/start-stop-restart.md) for the
change to take effect.

## Contributing

If you want to contribute to this project, we are humbled. Please take a look at
our [contributing guidelines](https://github.com/netdata/.github/blob/main/CONTRIBUTING.md) and don't hesitate to
contact us in our forums.

### How to develop a collector

Read [how to write a Netdata collector in Go](https://github.com/netdata/go.d.plugin/blob/master/docs/how-to-write-a-module.md).

## Troubleshooting

Plugin CLI:

```sh
Usage:
  orchestrator [OPTIONS] [update every]

Application Options:
  -m, --modules=    module name to run (default: all)
  -c, --config-dir= config dir to read
  -w, --watch-path= config path to watch
  -d, --debug       debug mode
  -v, --version     display the version and exit

Help Options:
  -h, --help        Show this help message
```

To debug specific module:

```sh
# become user netdata
sudo su -s /bin/bash netdata

# run plugin in debug mode
./go.d.plugin -d -m <module name>
```

Change `<module name>` to the module name you want to debug. See the [whole list](#available-modules) of available
modules.

## Netdata Community

This repository follows the Netdata Code of Conduct and is part of the Netdata Community.

- [Community Forums](https://community.netdata.cloud)
- [Netdata Code of Conduct](https://github.com/netdata/.github/blob/main/CODE_OF_CONDUCT.md)
