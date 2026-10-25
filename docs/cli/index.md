---
title: 'CLI'
order: 1
---

# <img src="/assets/cli.svg" style="display: inline;height: 1.5em;margin: 0;"> TAI CLI

TAI CLI is provided as a means to install custom variants of TA, and for use in applications that require programmatic control of the TA installation process, such as launchers.  
To use TAI CLI, download the portable `tai-cli.exe` executable from the [releases page](https://github.com/wilderzone/ta-installer/releases/latest) on GitHub.

Then run TAI, passing a [subcommand](/cli/subcommands), [options](/cli/options), and [flags](/cli/options#flags):
```sh
tai-cli.exe <subcommand> --repo=<url> --path=<path> -v
```

:::tip
This executable can be bundled with your application. See the [GUI package](https://github.com/wilderzone/ta-installer/tree/main/gui) for an example of this kind of integration.
:::

:::info
A [graphical version of TAI](/gui) is also provided in case you are not familiar with the command-line.
:::
