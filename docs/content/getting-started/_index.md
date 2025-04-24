---
title: "Getting Started"
weight: 1
---

## Installation

### Usign bash-wrapper

This will just download and install the latest binary from releases.

```bash
curl -sL https://raw.githubusercontent.com/floppa/yxa-install/main/install.sh | bash
```

Please note: Windows is not supported by this script.

### Using Go Install

```bash
go install github.com/floppa/yxa-cli@latest
```

Please note: That only running this command will give you "yxa-cli" instead of yxa
That can be fixed by renaming the binary.


### Download from release

#### Linux (amd64)

```bash
curl -L https://github.com/floppa/yxa-cli/releases/latest/download/yxa-linux-amd64 -o yxa
chmod +x yxa
sudo mv yxa /usr/local/bin/
```

#### macOS (Intel)

```bash
curl -L https://github.com/floppa/yxa-cli/releases/latest/download/yxa-darwin-amd64 -o yxa
chmod +x yxa
sudo mv yxa /usr/local/bin/
```

#### macOS (Apple Silicon)

```bash
curl -L https://github.com/floppa/yxa-cli/releases/latest/download/yxa-darwin-arm64 -o yxa
chmod +x yxa
sudo mv yxa /usr/local/bin/
```

#### Windows

Download the Windows executable from the assets below and add it to your PATH.

### Build from Source

```bash
git clone https://github.com/floppa/yxa-cli.git
cd yxa-cli
go build -o yxa
# Optional: move to a directory in your PATH
sudo mv yxa /usr/local/bin/