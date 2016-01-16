# Websocket hub for golang

[![GoDoc](https://godoc.org/github.com/DATA-DOG/golang-websocket-hub?status.svg)](https://godoc.org/github.com/DATA-DOG/golang-websocket-hub)

What? - ready to use websocket server with user subscription for private or broadcast messages.
It is based on [gorilla websockets](https://github.com/gorilla/websocket) package.

Why? - I find it a common use case for needing such behavior in day to day applications. Decided to opensource,
maybe someone will find it useful.

See the example application built with angular to show realtime behavior.

    cd example && make

Visit [localhost:8000](http://localhost:8000) to see it in action. Open a few tabs and try to send some messages.
