---
title: 'Tribes: Ascend Installer'
description: 'A CLI and GUI utility for managing Tribes: Ascend installations.'
---

<img src="/assets/banner.png" alt="">

# Overview

Tribes: Ascend Installer is a [CLI](https://en.wikipedia.org/wiki/Command-line_interface) and [GUI](https://en.wikipedia.org/wiki/Graphical_user_interface) utility for managing Tribes: Ascend installations on Windows and Linux.

These docs describe how to use TAI, how it works, and how to contribute to its development.


## Glossary

| Term           | Definition                                 |
| :------------- | :----------------------------------------- |
| Asset          | A downloadable collection of files, grouped by common functionality. Assets usually appear in the form of compressed archives. For example, all game files related to the Katabatic map are grouped into a single `Katabatic.zip` asset. |
| Manifest       | The machine-readable description of an installable entity (eg, the game). Manifests can reference one or more **assets**. |
