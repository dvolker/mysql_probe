# mysql probe

A golang application that checks a mysql database, writing results to flat files and/or runs an HTTP server for haproxy health checks based on those checks

## Installation

1. Install golang 1.3+
2. `go get github.com/haikulearning/mysql_probe`

### MySQL configuration

Create a MySQL user with the following permissions:
* PROCESS
* REPLICATION CLIENT

For example:

````sql
CREATE USER mysql_probe@'%' IDENTIFIED BY 'As92rj05UvKK';
GRANT PROCESS,REPLICATION CLIENT ON *.* TO mysql_probe@'%';
````

## Running

````
$ mysql_probe --help # will provide available commands & flags.
$ mysql_probe start  # will both test mysql and run the HTTP status server.
````

### Checks

For now, a default set of checks are hard-coded to run:
* "Connection Count Less Than X" with the following values:
  * 1, 10, 50, 100, 150, 200, 250, 300, 600, 1200, 2400
* "Replication Delay Less Than X" with the following values (in seconds):
  * 0, 10, 30, 60, 120, 300, 600, 1200, 2400

### haproxy Configuration Example
TOOD: Add an example haproxy config using these checks to build a resilient failover configuration.

### xinetd Configuration Example
TODO: Add an example of using xinetd to return these http txt files as http requests.

## Development

````bash
$ get go -u github.com/haikulearning/mysql_probe
$ cd $GOPATH/src/github.com/haikulearning/mysql_probe
$ make
````

## Testing
A dockerfile is available in docker/test for spinning up an instance of this app.

## Roadmap
1. Build checks with completely hardcoded configuration to check localhost.
2. Build configuration parser so this app does not need to be recompiled every time config changes.
3. Build rollup checks into configuration.

### Phase 1 hardcoded config
Check config.go for hardcoded values, host, username, password.
Run replcation_delay check and output to /var/run/mysql_probe/replcation_delay.http.txt
Run connection_count check and output to /var/run/mysql_probe/connection_count.http.txt

## Copyright

Copyright 2014 Derek Volker

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

