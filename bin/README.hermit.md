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
