# macOS

## Packaging

## Install the package

## Capturing System Audio

Linux gets the machine's output for free through Pulse's `.monitor` sources; macOS has no such thing

Running `make setup` will install [BlackHole](https://existential.audio/blackhole/)

```sh
brew install --cask blackhole-2ch

# CoreAudio only picks the new driver up after a reboot or
sudo killall coreaudiod
```

Then open **Audio MIDI Setup** app, create a Multi-Output Device combining your speakers and BlackHole 2ch, and select it as the system output, then pick `BlackHole 2ch` in Ekko. Without that, only the microphone devices are usable.

The first run prompts for microphone access; grant it to the terminal (`make dev`) or to Ekko.app (`make build`), otherwise ffmpeg records silence.
