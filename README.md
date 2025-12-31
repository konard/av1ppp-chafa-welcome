<p align="center">

<img src="./img/header.jpg" />
<br />
<br />

<a href="./LICENSE">
<img src="https://img.shields.io/badge/license-blue?style=for-the-badge" alt="License" />
</a>
&emsp;
<img src="https://img.shields.io/badge/platform-linux-blue?style=for-the-badge&logo=linux" alt="Linux" />
<br />

</p>

<p align="center">
Command line tool for getting system information with a render via
<a href="https://github.com/hpjansson/chafa">chafa</a>
</p>

---

## Pre-requirements

1. First you need to install chafa:
    - The current version can always be installed from
      [the repository](https://github.com/hpjansson/chafa/tree/master#installing).
    - Arch / Manjaro:  
      `sudo pacman -S chafa`
    - Ubuntu / Debian / Mint:  
      `sudo apt install chafa`

2. You also need to have [golang](https://go.dev/) installed

3. Make sure that the `~/go/bin` directory is added to the `PATH`

## Download and install

```bash
go install github.com/av1ppp/chafa-welcome/cmd/chafa-welcome@latest
chafa-welcome
```

If the installation was successful, you will get an error:

```
panic: validation error: stat /path/to/image.jpg: no such file or directory
```

You will only need to specify the path to the image in the
`~/.chafa-welcome/config` file in the `source` field:

```toml
[image]
source = '/real/path/to/image.jpg'
```

## Community Features

### Browse Community Presets

Discover and apply configurations shared by the community:

```bash
chafa-welcome --gallery
# or
chafa-welcome --community
```

This opens an interactive browser where you can:
- Browse popular, newest, or alphabetically sorted presets
- Filter by category (e.g., "Cyberpunk", "Minimal Dark", "Retro Green")
- Preview preset details and apply with a single key press
- Presets are automatically downloaded and configured

### Share Your Configuration

Share your current configuration with others via GitHub Gist:

```bash
# Share as private gist (requires gh CLI or GITHUB_TOKEN)
chafa-welcome --share

# Share as public gist
chafa-welcome --share --public

# Include your ASCII art image
chafa-welcome --share --with-image
```

### Import Configuration

Import a configuration from a shared GitHub Gist:

```bash
chafa-welcome --import https://gist.github.com/username/gist_id
```

### Contributing to the Gallery

To add your preset to the community gallery:

1. Share your configuration: `chafa-welcome --share --public`
2. Create a Pull Request at [chafa-welcome-presets](https://github.com/av1ppp/chafa-welcome-presets)

## Command Line Options

```
Usage: chafa-welcome [OPTIONS]

Options:
  --help            Show help message
  --version         Show version information

Community Features:
  --gallery         Browse community presets gallery
  --community       Alias for --gallery

Sharing:
  --share           Share current configuration to GitHub Gist
  --public          Make the shared gist public (default: private)
  --with-image      Include the image when sharing
  --import URL      Import configuration from a GitHub Gist URL
```

## Development environment

The application was developed and tested with the following versions:

- golang - 1.20.4
- chafa - 1.13.0


