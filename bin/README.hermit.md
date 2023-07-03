# Installation

Install it with the following commands:
```zsh
curl -fsSL https://github.com/cashapp/hermit/releases/download/stable/install.sh | /bin/bash
. ./bin/activate-hermit
```
> **Note**
> This will download and install hermit into `~/bin`. You should add this to your `$PATH` if it isn't already.
It is also recommended to add hermit's [shell integration](https://cashapp.github.io/hermit/usage/shell/)


# Hermit environment

This is a [Hermit](https://github.com/cashapp/hermit) bin directory.

The symlinks in this directory are managed by Hermit and will automatically
download and install Hermit itself as well as packages. These packages are
local to this environment.

Want to update a pacakge?
```shell
hermit install opa@latest # updates to the latest version
```
or
```shell
hermit install opa@0.46 # install a specific version
```

