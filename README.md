# [WIP] Caronte

[![Build Status](https://travis-ci.com/eciavatta/caronte.svg?branch=develop)](https://travis-ci.com/eciavatta/caronte)
[![codecov](https://codecov.io/gh/eciavatta/caronte/branch/develop/graph/badge.svg)](https://codecov.io/gh/eciavatta/caronte)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/009dca44f4da4118a20aed2b9b7610c0)](https://www.codacy.com/manual/eciavatta/caronte?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=eciavatta/caronte&amp;utm_campaign=Badge_Grade)

<img align="left" src="https://divinacommedia.weebly.com/uploads/5/5/2/3/5523249/1299707879.jpg">
Caronte is a tool to analyze the network flow during capture the flag events of type attack/defence.
It reassembles TCP packets captured in pcap files to rebuild TCP connections, and analyzes each connection to find user-defined patterns.
The patterns can be defined as regex or using protocol specific rules.
The connection flows are saved into a database and can be visualized with the web application. REST API are also provided.

## Installation
There are two ways to install Caronte:
- with Docker and docker-compose, the fastest and easiest way
- manually installing dependencies and compiling the project

### Run with Docker
The only things to do are:
- clone the repo, with `git clone https://github.com/eciavatta/caronte.git`
- inside the `caronte` folder, run `docker-compose up --build -d`
- wait for the image to be compiled and open browser at `http://localhost:3333`

### Manually installation
The first thing to do is to install the dependencies:
- go >= 1.14 (https://golang.org/doc/install) 
- node >= v12 (https://nodejs.org/it/download/)
- yarnpkg (https://classic.yarnpkg.com/en/docs/install/)
- hyperscan >= v5 (https://www.hyperscan.io/downloads/)

Next you need to compile the project, which is composed of two parts:
- the backend, which can be compiled with `go mod download && go build`
- the frontend, which can be compiled with `cd frontend && yarn install && yarn build`

Before running Caronte starts an instance of MongoDB (https://docs.mongodb.com/manual/administration/install-community/) that has no authentication. _Be careful not to expose the MongoDB port on the public interface._

Run the binary with `./caronte`. The available configuration options are:
```
-bind-address    address where server is bind (default "0.0.0.0")
-bind-port       port where server is bind (default 3333)
-db-name         name of database to use (default "caronte")
-mongo-host      address of MongoDB (default "localhost")
-mongo-port      port of MongoDB (default 27017)
```

## Configuration
The configuration takes place at runtime on the first start via the graphical interface (TO BE IMPLEMENTED) or via API. It is necessary to setup:
- the `server_address`: the ip address of the vulnerable machine. Must be the destination address of all the connections in the pcaps. If each vulnerable service has an own ip, this param accept also a CIDR address. The address can be either IPv4 both IPv6
- the `flag_regex`: the regular expression that matches a flag. Usually provided on the competition rules page
- `auth_required`: if true a basic authentication is enabled to protect the analyzer
- an optional `accounts` array, which contains the credentials of authorized users
