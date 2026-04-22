# IFLy installation guide

This guide covers building the `ifly` CLI from source and installing the Claude
Code plugin.

## Prerequisites

- Claude Code installed and logged in.
- Git.
- Go.
- Bash and `jq` for the plugin hooks.

On Windows, Git Bash is the easiest way to provide Bash. On Linux/macOS, Bash is
normally already present.

## Windows

Install Git, Go, and jq with winget:

```powershell
winget install --id Git.Git -e
winget install --id GoLang.Go -e
winget install --id jqlang.jq -e
```

Close and reopen your terminal after installing. This matters because `go`,
`git`, `bash`, and `jq` are added to `PATH` by the installers.

Verify:

```powershell
go version
git --version
bash --version
jq --version
```

Clone and build:

```powershell
git clone https://github.com/ljn7/ifly.git
cd ifly
.\build-cli.bat
```

Run the CLI:

```powershell
.\dist\ifly.exe version
.\dist\ifly.exe paths
```

Install the Claude plugin:

```powershell
.\dist\ifly.exe install
```

If you already installed it before:

```powershell
.\dist\ifly.exe install --overwrite
```

Then in Claude Code:

```text
/reload-plugins
/ifly:status
```

Optional: put `ifly.exe` on your `PATH`.

```powershell
New-Item -ItemType Directory -Force "$env:USERPROFILE\bin"
Copy-Item .\dist\ifly.exe "$env:USERPROFILE\bin\ifly.exe" -Force
```

Add `%USERPROFILE%\bin` to your user `PATH`, then restart the terminal.

## Ubuntu/Debian

Install dependencies:

```bash
sudo apt update
sudo apt install -y git golang-go jq bash make tar
```

Clone and build:

```bash
git clone https://github.com/ljn7/ifly.git
cd ifly
./build-cli.sh
```

Run:

```bash
./dist/ifly version
./dist/ifly paths
```

Install the plugin:

```bash
./dist/ifly install
```

Optional: install the binary to your user bin directory:

```bash
mkdir -p ~/.local/bin
install -m 0755 ./dist/ifly ~/.local/bin/ifly
```

Make sure `~/.local/bin` is on `PATH`:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
exec bash
```

## Arch Linux

Install dependencies:

```bash
sudo pacman -Syu --needed git go jq bash make tar
```

Clone, build, and install:

```bash
git clone https://github.com/ljn7/ifly.git
cd ifly
./build-cli.sh
./dist/ifly install
```

Optional user-bin install:

```bash
mkdir -p ~/.local/bin
install -m 0755 ./dist/ifly ~/.local/bin/ifly
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
exec bash
```

## Fedora/RHEL

Install dependencies:

```bash
sudo dnf install -y git golang jq bash make tar
```

Build and install:

```bash
git clone https://github.com/ljn7/ifly.git
cd ifly
./build-cli.sh
./dist/ifly install
```

## macOS

Install dependencies with Homebrew:

```bash
brew install go git jq
```

Build and install:

```bash
git clone https://github.com/ljn7/ifly.git
cd ifly
./build-cli.sh
./dist/ifly install
```

Optional user-bin install:

```bash
mkdir -p ~/.local/bin
install -m 0755 ./dist/ifly ~/.local/bin/ifly
```

## First-run setup

Set a default mode and guard:

```bash
ifly mode minimal
ifly guard strict
ifly status
```

For one project only:

```bash
cd /path/to/project
ifly mode --project normal
ifly guard --project project
```

Add common safety presets:

```bash
ifly block preset git-danger
ifly block preset filesystem-danger
```

Inside Claude Code:

```text
/reload-plugins
/ifly:status
```

## Updating

From a source checkout:

```bash
git pull
./build-cli.sh
./dist/ifly install --overwrite
```

On Windows:

```powershell
git pull
.\build-cli.bat
.\dist\ifly.exe install --overwrite
```

Reload Claude plugins after updating:

```text
/reload-plugins
```

## Troubleshooting

Check where IFLy thinks Claude stores plugins:

```bash
ifly paths
```

If Claude does not see `/ifly:status`:

```text
/reload-plugins
```

If Windows says `bash` or `jq` is not found, restart the terminal. If it still
fails, verify Git and jq are on `PATH`:

```powershell
where.exe bash
where.exe jq
```

If global state and Claude state disagree on Windows, make sure you are running
the current plugin hooks and then reload:

```powershell
.\dist\ifly.exe install --overwrite
```

```text
/reload-plugins
```
