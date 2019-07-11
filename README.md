# Ambilight

This project aims to deliver a robust client-server command-line application that is able to stream live, low-latency LED color data to a controller via a wireless / wired network connection.

# /!\ WIP /!\

The project is currently under development. Features may be missing, will generally be unstable and setup/build information are missing and will be added in the near future.

**TODO:**
- [ ] Setup instructions
- [ ] Build instructions
- [ ] Add dependencies on README
- [ ] Documentation
- [ ] Contributing instructions
- [ ] Add binaries for Windows / Linux

**PROGRAMMING TODO:**
- [x] Include more modes (Rainbow, Pulse etc.)
- [ ] Proof of concept app for controlling using a smartphone



## How it works

The client is launched on a Windows / Mac / Linux PC and the server on a controller (e.g. Raspberry Pi) that is wired with the LED strip that will be controlled.

Once the server is running on the controller, whenever the client is launched it will try to connect with the command-line parameters provided.

Once the connection is established, the client will stream data to the server via a TCP socket, and the server will act according to the specified mode (for now only Render and Rainbow modes are supported, more will be added soon!)

## Modes

```txt
A : Ambilight : Renders the streaming LED data that are provided from the client.
R : Rainbow   : Infinite loop of a gradient color shift animation.
```

## Under the hood

This library consists of 3 packages, a client (client.go), a server (server.go), and a wrapper package (ambilight.go) that has functions that the client and server use to connect to one another and transmit/receive data.
There are verbose comments and documentation throughout these packages, detailing how everything is set up.

When the socket connection is established, the client sends a TCP packet with this format:


```
Bytes (hex):

  - 1: MMMM MMMM * M is mode ascii character
  - 2: RRRR RRRR * R is red
  - 3: GGGG GGGG * G is green
  - 4: BBBB BBBB * B is blue
  - 5: RRRR RRRR * repeats for each additional LED
  - 6: GGGG GGGG
  - 7: BBBB BBBB
       ...
```

The first byte is the ASCII **mode** character in binary.
<br>
The rest of the bytes that follow MUST be `N * 3`, where `N` is the number of LEDs that will be controlled.
<br>
If the strip has more leds the rest of the LEDs' behavior is undefined, if it has less the underlying ws2811.c library will probably error out (haven't tested that out, ymmv)


## How it works

The **client** is launched on a Windows / Mac / Linux PC and the **server** on a controller (e.g. Raspberry Pi) that is wired with the LED strip that will be controlled.

Once the server is running on the controller, whenever the client is launched it will try to connect with the command-line parameters provided.

Once the connection is established, the client will stream data to the server via a TCP socket, and the server will act according to the specified mode of operation (for now only Render and Rainbow modes are supported, more will be added soon!)
