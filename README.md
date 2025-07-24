# LAN Drop

> A simple Go GUI with an embedded HTTP server that allows multiple file upload.

## The problem

Coming from the magnificent (and this is sarcastic!) Apple Ecosystem, I found myself in pain having to transfer all sorts of files from my iPhone
to my Windows Desktop.

So I came up with this simple yet powerful solution, using HTTP to _emulate_ the AirDrop feature.

## Basic Instructions

1. Install Go compiler and setup PATH variables (this is crucial!!!!)
2. With your phone (or whatever) scan the QR code.
3. Select the files you want to upload.
4. Click Upload and the Job is done.

Almost as easy as Apple's right?

## Installing

> If you find yourself having trouble with the process please contact me.

Navigate to the latest release and download the zip folder containing the version compatible with your os

## Security Concerns

This app is currently under development, and I'm planning to add a sort of encryption layer, but right now it's very likely to be vulnerable to spoofing attacks via the http protocol.
