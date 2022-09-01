# ledctl

This project aims to deliver a robust client-server command-line application
that is able to stream live, low-latency LED color data to a controller via a
wireless / wired network connection.

---

## Work in progress

The project is currently under development. There might be bugs or inaccuracies
on some parts of the documentation.

---

## How it works

The *client* is launched on a Windows machine (Mac and Linux coming soon™) and
the *server* on a microcontroller (e.g. Raspberry Pi) wired with the LED strip.

Once the server is running on the microcontroller, whenever the client is
launched it will try to connect with the connection details provided through
a config file.

When the connection is established, the client will start streaming data to the
server via websocket, and the server will react according to the specified mode
of operation (for now only the 'ambilight' mode is supported, more will be
added soon).

<br>

*So how do you install it?*

---

## Installation instructions

#### Client (e.g. Windows PC)

1. Download the *client* binary for your operating system from
[here](https://github.com/rdnt/ledctl/releases/latest/). Place it anywhere
you want. Launch the executable. Right-click on the tray icon and quit.

2. The default `ledctl.json` file was just created in your working directory.

3. Edit the file to match your setup. **NOTE:** `server.leds` should be the
sum of all of your displays' LEDs.

4. Launch the client again. It will automatically connect to the server once
the server is online.


#### Server (e.g. Raspberry Pi Zero W 2)

0. Prerequisite: Have SSH enabled and be connected to a wifi network.  
For setup information look here:
[Headless Pi Zero W Wifi Setup (Windows)](https://desertbot.io/blog/headless-pi-zero-w-wifi-setup-windows)

1. Login to the Raspberry Pi via SSH.

2. Install `tmux` using the following command:

  `sudo apt-get install tmux -y`

3. Download the server binary (for example using `wget`):

  `wget https://github.com/rdnt/ledctl/releases/download/0.0.1-pre-release/ledctld-linux-arm64`

4. Mark the binary as executable:

  `sudo chmod +x ledctld-linux-arm64`

8. To start the server, simply write `sudo ./ledctld-linux-arm64` on the terminal.

9. *(optional)* Start the server at boot:  
Edit the `/etc/rc.local` file (`sudo nano /etc/rc.local`), adding the following
before the `exit 0` line, replacing `LEDCTL_DIR` with the directory where the
ledctld binary resides.

  `tmux new-session -d -s ledctl 'cd /LEDCTL_DIR && sudo ./ledctld-linux-arm64'`

  If you reboot the server will start automatically.

---

## Modes

```txt
Ambilight - Video : Captures the screen and sends averaged color data for each
  frame and for each monitor to its respecting LED segment.
Ambilight - Audio: same as Video, but captures audio and sends an audio spectrum
  instead.
```

---

## Dependencies

This project depends on the following libraries, among others:

- [gadgetoid, supcik, urmel11](https://github.com/orgs/rpi-ws281x/people) /
  [rpi_ws281x](https://github.com/rpi-ws281x/rpi-ws281x-go) -
  Raspberry Pi library for controlling WS281X LEDs
- [getlantern](https://github.com/getlantern) /
  [systray](github.com/getlantern/systray) -
  A library providing an easy API to add tray functionality

---

## Contributing
You are free and actively encouraged to contribute to this project by either
contributing code, creating issues, reporting bugs, highlighting
vulnerabilities, proposing improvements or helping maintain the documentation.

If you would like to submit code changes, create a new branch from the *main*
branch with the name of the feature you are implementing and submit a pull
request to the *main* branch after you make your changes. Click
[here](https://gist.github.com/Chaser324/ce0505fbed06b947d962#doing-your-work)
for a how-to guide.

In case you want to submit a bug report, please add as many details as possible
regarding how the error occurred and include the steps required to reproduce
it if that is possible. It will help a lot in testing, finding the cause and
implementing fixes.

---

## Changelogs
Changelogs for each and every release can be found
[here](https://github.com/rdnt/ledctl/releases).

---

## Copyright
Any reproductions of this project *must* include a link to this repository and
the project's LICENSE file.
