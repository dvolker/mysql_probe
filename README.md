mysql probe
==========
golang application that checks a mysql database and creates a series of HTTP responses stored as
flat text files that xinetd can return quickly for haproxy health checks.

## Installation
Install golang 1.3+ and run ./build

## Create package for distribution
Install fpm and run ./package https://github.com/jordansissel/fpm

## Configuration
Edit /etc/mysqlprobe/config.yaml

## Running
mysql_probe --help will provide available options.

## Checks
Checks can be named and run at configured intervals. Each check can 'roll up' multiple checks and
return an alarm if any of the checks fail.
### Global check options
These options apply to an entire instance of mysql_probe.
host: hostname to connect to
user: user to login to mysql with
pass: password to login to mysql with
interval: seconds to wait before rerunning this check. This is calculated as starting when the previous check has finished to prevent pileups.

### Individual check options
name: display name of this check. must be unique.
type: required check name
result_files: a list of files to place this result as http responses in text

### replication_delay
Checks that replication is not delayed over configured seconds.
#### Options
max_delay: in seconds, if replication delayed > this number, consider check failed.

### connection_count
Count the existing connections to this server and fail if connection count is too great.
#### Options
max_connections: if there are more than this many connections to the server, fail

### assert
Run a query that returns a number and compare the results
#### Options
query: query to run
fail_if_gt: compare the numeric result of query and fail if the returned value is greater than this number.
fail_if_lt: compare the numeric result of query and fail if the returned value is less than this number.

## Development phases
1. Build checks with completely hardcoded configuration to check localhost.
2. Build configuration parser so this app does not need to be recompiled every time config changes.
3. Build rollup checks into configuration.

## Phase 1 hardcoded config
Check config.go for hardcoded values, host, username, password.
Run replcation_delay check and output to /var/run/mysql_probe/replcation_delay.http.txt
Run connection_count check and output to /var/run/mysql_probe/connection_count.http.txt

## Testing
A dockerfile is available in docker/test for spinning up an instance of this app.

## haproxy Configuration Example
TOOD: Add an example haproxy config using these checks to build a resilient failover configuration.

## xinetd Configuration Example
TODO: Add an example of using xinetd to return these http txt files as http requests.

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

