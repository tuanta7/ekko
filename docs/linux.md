# Linux (Ubuntu, Arch)

## Packaging

Build the binary, then create a package in `bin/`:

```sh
make build
wails3 tool package -name ekko -format deb -config ./build/linux/nfpm/nfpm.yaml -out ./bin
```

Swap `-format deb` for `rpm` or `archlinux`.

## Install the package

The package ships the binary, the model, the icon and the `.desktop` entry; GTK, WebKit, `ffmpeg` and `pulseaudio-utils` are pulled in as dependencies.

```sh
sudo apt install ./bin/ekko.deb
sudo apt remove ekko
```
