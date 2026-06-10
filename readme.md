# Go Live Server

A lightweight Live Server CLI for frontend developers, written in Go.

> **Status:** In active development.

## Getting Started

### Prerequisites

- **Pre-compiled binaries:** None required.
- **Build from source:** Go 1.26+.

### Installation

#### Option 1: Pre-compiled Binary

Download the latest binary for your OS (Windows, macOS, Linux) from the [Releases](https://github.com/phmshk/go-live-server/releases) page.

1. Download the archive for your OS.
2. Extract the binary.
3. (Optional) Move it to your `PATH`.

#### Option 2: Build from Source

```bash
git clone https://github.com/phmshk/go-live-server.git
cd go-live-server
go build .
```

To compile with a custom filename:

```bash
go build -o gls .
```

### Usage

Run the binary from the command line.

| Flag            | Description                                                  |
| --------------- | ------------------------------------------------------------ |
| `--port=NUMBER` | Port to serve on (default: `8080`). Auto-increments if busy. |
| `--dir=PATH`    | Directory to watch (default: current directory).             |
