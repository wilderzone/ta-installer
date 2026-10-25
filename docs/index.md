<img src="/assets/banner.png" alt="" style="border-radius: 8px;">

# Overview

Tribes: Ascend Installer is a [CLI](https://en.wikipedia.org/wiki/Command-line_interface) and [GUI](https://en.wikipedia.org/wiki/Graphical_user_interface) utility for managing Tribes: Ascend installations on Windows and Linux.

These docs describe how to use TAI, how it works, and how to contribute to its development.


## Table of Contents

- [Glossary](#glossary)
- [Command-Line Interface](/cli)
  - [Subcommands](/cli/subcommands)
  - [Options](/cli/options)
  - [Output](/cli/output)
  - [Dependency Resolution](/cli/dependency-resolution)
- [Graphical User Interface](/gui)
  - [Game Type](/gui/game-type)
  - [Location](/gui/location)
  - [Action](/gui/action)
- [Development](/development)
  - [Project Structure](/development/project-structure)
  - [Running on Linux](/development/running-on-linux)
  - [Installation Utilities](/development/installation-utilities)
  - [Building](/development/building)


## Glossary

| Term           | Meaning                                    |
| :------------- | :----------------------------------------- |
| Asset          | A downloadable collection of files, grouped by common functionality. Assets usually appear in the form of compressed archives. For example, all game files related to the Katabatic map are grouped into a single `Katabatic.zip` asset. |
| Manifest       | The machine-readable description of an installable entity (eg, the game). Manifests can reference one or more **assets**. |
