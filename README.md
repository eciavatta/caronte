# [WIP] Caronte

[![Build Status](https://travis-ci.com/eciavatta/caronte.svg?branch=develop)](https://travis-ci.com/eciavatta/caronte)
[![codecov](https://codecov.io/gh/eciavatta/caronte/branch/develop/graph/badge.svg)](https://codecov.io/gh/eciavatta/caronte)
[![GPL License][license-shield]][license-url]

<img align="left" src="https://divinacommedia.weebly.com/uploads/5/5/2/3/5523249/1299707879.jpg">
Caronte is a tool to analyze the network flow during capture the flag events of type attack/defence.
It reassembles TCP packets captured in pcap files to rebuild TCP connections, and analyzes each connection to find user-defined patterns.
The patterns can be defined as regex or using protocol specific rules.
The connection flows are saved into a database and can be visualized with the web application. REST API are also provided.

