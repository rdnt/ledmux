# Ambilight-Controller

This project aims to deliver a robust client-server command-line application that is able to stream live, low-latency LED data to a controller via a wireless / wired network connection.

# /!\ WIP

The project is currently under development. Features may be missing, will generally be unstable and setup/build information are missing and will be added in the near future.

**TODO:**
 - Setup instructions
 - Build instructions
 - Add dependencies on README
 - Documentation
 - Contributing instructions
 - Add binaries for Windows / Linux
 - Proof of concept app for controlling using a smartphone

## How it works

The <u>client</u> is launched on a Windows / Mac / Linux PC and the <u>server</u> on a controller (e.g. Raspberry Pi) that is wired with the LED strip.

Once the server is running on the controller, whenever the client is launched it will try to connect with the command-line parameters provided.

Once the connection is established, the client will stream data to the server via a TCP socket, and the server will act depending on the specified mode of operation (for now only Render mode is supported, more will be added soon!)

## Operation Modes

```txt
A : Ambilight - Renders the streaming LED data that are provided from the server.
```
