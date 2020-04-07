# [WIP] Caronte

[![Build Status](https://travis-ci.com/eciavatta/caronte.svg?branch=develop)](https://travis-ci.com/eciavatta/caronte)
[![codecov](https://codecov.io/gh/eciavatta/caronte/branch/develop/graph/badge.svg)](https://codecov.io/gh/eciavatta/caronte)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/009dca44f4da4118a20aed2b9b7610c0)](https://www.codacy.com/manual/eciavatta/caronte?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=eciavatta/caronte&amp;utm_campaign=Badge_Grade)

<img align="left" src="https://divinacommedia.weebly.com/uploads/5/5/2/3/5523249/1299707879.jpg">
Caronte is a tool to analyze the network flow during capture the flag events of type attack/defence.
It reassembles TCP packets captured in pcap files to rebuild TCP connections, and analyzes each connection to find user-defined patterns.
The patterns can be defined as regex or using protocol specific rules.
The connection flows are saved into a database and can be visualized with the web application. REST API are also provided.
