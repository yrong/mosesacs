# Moses ACS [![Build Status](https://travis-ci.org/lucacervasio/mosesacs.svg?branch=master)](https://travis-ci.org/lucacervasio/mosesacs)

An ACS in Go for provisioning CPEs, suitable for test purposes or production deployment.

## Getting started

Install the package:

    go get github.com/lucacervasio/mosesacs

Run daemon:

    mosesacs -d

Connect to it and get a cli:

    mosesacs

Congratulations, you've connected to the daemon via websocket. Now you can issue commands via cli o browse the embedded webserver at http://localhost:9292/www

## Services exposed

Moses exposes three services:

 - http://localhost:9292/acs is the endpoint for the CPEs to connect
 - http://localhost:9292/www is the embedded webserver to control your CPEs
 - ws://localhost:9292/ws is the websocket endpoint used by the cli to issue commands. Read about the API specification if you want to build a custom frontend which interacts with mosesacs daemon.


