/*
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
*/

package main

import (
	"database/sql"
	"fmt"
	"github.com/codegangsta/cli"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
)

func main() {
	// TODO: parse flags
	app := cli.NewApp()
	app.Name = "mysql_probe"
	app.Usage = "test mysql health and write out http txt responses"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "host",
			Value:  "127.0.0.1",
			Usage:  "mysql host to connect to",
			EnvVar: "MYSQL_PROBE_HOST",
		},
		cli.IntFlag{
			Name:   "port, p",
			Value:  3306,
			Usage:  "mysql port to connect to",
			EnvVar: "MYSQL_PROBE_PORT",
		},
		cli.StringFlag{
			Name:   "user, u",
			Value:  "root",
			Usage:  "mysql username to connect with",
			EnvVar: "MYSQL_PROBE_USER",
		},
		cli.StringFlag{
			Name:   "pass",
			Value:  "test",
			Usage:  "mysql password to connect with",
			EnvVar: "MYSQL_PROBE_PASS",
		},
		cli.IntFlag{
			Name:   "interval, i",
			Value:  1,
			Usage:  "interval in seconds to run the checks",
			EnvVar: "MYSQL_PROBE_INTERVAL",
		},
	}
	app.EnableBashCompletion = true
	app.Action = func(c *cli.Context) {
		// setup checks to run on intervals
		initTest(c.String("host"), c.Int("port"), c.String("user"), c.String("pass"), c.Int("interval"))
	}
	app.Run(os.Args)

	// TODO: override hard coded config with env vars then cmd line flags
	// TODO: output logs in logstash json format
	// TODO: write output files as failures on termination
}

func initTest(host string, port int, user string, pass string, interval int) {
	if interval <= 0 {
		panic("interval must be a positive integer")
	}
	// run checks on intervals
	for _ = range time.Tick(time.Duration(interval) * time.Second) {
		test(host, port, user, pass)
	}
}
func test(host string, port int, user string, pass string) {
	fmt.Println("Host: ", host, " Port: ", port, " User: ", user, " Pass: ", pass)

	// Create dsn like such https://github.com/Go-SQL-Driver/MySQL/#dsn-data-source-name
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	// username:password@protocol(address)/dbname?param=value
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", user, pass, host, port)
	fmt.Println("DSN: ", dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	fmt.Println("SUCCESS")
}
