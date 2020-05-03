# pmjpg: Proxy mjpg

pmjpg is kinda proxy for a mjpg camera.

# Why

This was needed to be able to use a IP camera (DSC-5030L) inside a Diapora (PowerPoint).
A plugin is used to show a web page in the Diapora (www.liveslides.com) which embed a
Chromium version 43.0.2357.0 which doesn't support the mjpg format.

An other problem needed to be resolved: the URL `http://user:pwd@ip` is not supported anymore,
but it was mandatory that the camera couldn't be used without some kind of credential.

# Use

It's a simple GO HTTP sever, which will exposed a `GET /image` handler and any other one define
in the template directory.

The template can be any valid HTML page with some variable to be define, and which should call
the `Get /image` at the desired frequency, see the `idx?.html` file for some example.

The "X-Idx" optional header is used to reused the same connection to the camera allowing a much
faster frequency and avoiding sharing it with different client. It's the server that will
generate it, when generating the template.
