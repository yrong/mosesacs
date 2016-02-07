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

### 1. `list`: list CPEs

 example:

```
 moses@localhost:9292/> list
 cpe list
 CPE A54FD with OUI 006754
```

### 2. `readMib serial leaf/subtree`: read a specific leaf or a subtree

 example:

```
 moses@localhost:9292/> readMib A54FD Device.
 Received an Inform from [::1]:58582 (3191 bytes) with SerialNumber A54FD and EventCodes 6 CONNECTION REQUEST
 InternetGatewayDevice.Time.NTPServer1 : pool.ntp.org
 InternetGatewayDevice.Time.CurrentLocalTime : 2014-07-11T09:08:25
 InternetGatewayDevice.Time.LocalTimeZone : +00:00
 InternetGatewayDevice.Time.LocalTimeZoneName : Greenwich Mean Time : Dublin
 InternetGatewayDevice.Time.DaylightSavingsUsed : 0
```


## Services exposed

Moses exposes three services:

 - http://localhost:9292/acs is the endpoint for the CPEs to connect
 - http://localhost:9292/www is the embedded webserver to control your CPEs
 - ws://localhost:9292/ws is the websocket endpoint used by the cli to issue commands. Read about the API specification if you want to build a custom frontend which interacts with mosesacs daemon.


