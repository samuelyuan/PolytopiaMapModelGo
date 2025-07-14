# Polytopia Map Model Go

This is a library aimed at reading state files and creating a detailed and interactive map model for the game The Battle of Polytopia. This project leverages Go programming language to build and manage the map data.

## Features

- Detailed terrain and resource representation
- Contains all units and improvement data
- Customizable map settings
- Efficient data management with Go

## Supported File Formats

This library supports parsing `.state` files used by The Battle of Polytopia.

- [`.state` file format spec](docs/state_format.md) – a compressed, all-in-one save file containing map layout, units, players, game settings, and action history.

These specifications define how the binary file structure is mapped to Go structs in this repository.
