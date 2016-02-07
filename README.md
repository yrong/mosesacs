# Moses ACS [![Build Status](https://travis-ci.org/lucacervasio/mosesacs.svg?branch=master)](https://travis-ci.org/lucacervasio/mosesacs)

An ACS in Go for provisioning CPEs, suitable for test purposes or production deployment.

## Getting started

Install the package:

    go get github.com/lucacervasio/mosesacs

Run daemon:

    mosesacs -d

Connect to it and get a cli:

    mosesacs

Congratulations, you've connected to the daemon via websocket. Now you can issue commands via CLI or browse the embedded webserver at http://localhost:9292/www

## Compatibility on ARM

Moses is built on purpose only with dependencies in pure GO. So it runs on ARM processors with no issues. We tested it on QNAP devices and Raspberry for remote control.

## CLI commands

 ### `list`: list CPEs

 ### `readMib`: read a specific leaf or a subtree

 `readMib serial leaf/subtree`

 example:

 moses@localhost:9292/> list
 elenco cpe
 CPE A54FD with OUI 006754
 moses@localhost:9292/> readMib A54FD Device.


## Services exposed

Moses exposes three services:

 - http://localhost:9292/acs is the endpoint for the CPEs to connect
 - http://localhost:9292/www is the embedded webserver to control your CPEs
 - ws://localhost:9292/ws is the websocket endpoint used by the cli to issue commands. Read about the API specification if you want to build a custom frontend which interacts with mosesacs daemon.


