# ModrinthCLI

My first Go project

A command-line tool to manage Minecraft mods, resource packs, and shaders using the Modrinth API.
Works regardless of if its a .minecraft or not, so mod installer and updater should work for servers, too!
I may port this to a minecraft server manager because I have been looking for one that is similar to Prism

## Features

- Add, enable, disable, and list mods
- Install and update mods from Modrinth
- Supports multiple loaders (fabric, neoforge, etc.)
- Manifest-based project management

## Usage

```sh
mod init [DIRECTORY]           # Create a new project manifest
mod add <slug> [--to DIR]      # Add a mod by slug
mod list                       # List all manifest entries
mod enable <slug> [...]        # Enable mods
mod disable <slug> [...]       # Disable mods
mod install                    # Download/install enabled mods
mod update [--dry-run]         # Check for and install updates
```

See `mod <command> --help` for more options.

## TODO

- [x] Add support for removing mods from the manifest
- [x] Add search TODO: resourcepacks, shaderpacks
- [ ] Improve error messages and logging
- [ ] Support for curseforge
- [ ] Support for custom jars and resource packs?
- [ ] Unit tests
- [x] Actually make resourcepacks and shader installation work
- [ ] Windows
- [ ] Fuzzy Search for enabling/disabling/removing items
---

MIT License or something

## Installation

### Linux

Unarchive the ``mod_vX.X.X_linux_x86_64.tar.gz``
Add ``mod`` to your path or just add it in a directory and use ``./mod (command)``

### Mac

Unarchive the ``mod_vX.X.X_linux_x86_64.tar.gz``
Add ``mod`` to your path or just add it in a directory and use ``./mod (command)``

### Windows

TODO