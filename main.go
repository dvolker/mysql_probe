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
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/dvolker/mysql_probe/mysqltest"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

const VERSION string = "0.0.2"

func main() {
	// TODO: parse flags
	app := cli.NewApp()
	app.Name = "mysql_probe"
	app.Usage = "test mysql health and write out http txt responses"
	app.Version = VERSION
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
		cli.StringFlag{
			Name:   "jsonlog",
			Value:  "/dev/stdout",
			Usage:  "file to log output in json",
			EnvVar: "MYSQL_PROBE_JSONLOG",
		},
		cli.StringFlag{
			Name:   "reports",
			Value:  "tmp",
			Usage:  "directory to write reports to",
			EnvVar: "MYSQL_PROBE_REPORTS",
		},
		cli.IntFlag{
			Name:   "interval, i",
			Value:  250,
			Usage:  "interval in milliseconds to run the checks, set to 0 to only run the tests once",
			EnvVar: "MYSQL_PROBE_INTERVAL",
		},
		cli.IntFlag{
			Name:   "timeout, t",
			Value:  2000,
			Usage:  "time in milliseconds to wait for a mysql connection",
			EnvVar: "MYSQL_PROBE_TIMEOUT",
		},
	}
	app.EnableBashCompletion = true
	app.Action = func(c *cli.Context) {
		// setup checks to run on intervals

		file, err := os.OpenFile(c.String("jsonlog"), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			panic(fmt.Sprintf("Couldn't open jsonlog \"%s\" for writing: %s", c.String("jsonlog"), err.Error()))
		}
		defer file.Close()
		t := mysqltest.NewMysqlTest("connection", c.String("host"), c.Int("port"), c.String("user"), c.String("pass"), c.Int("interval"), c.Int("timeout"), c.String("reports"), file)
		t.Run()
	}
	app.Run(os.Args)

	// TODO: override hard coded config with env vars then cmd line flags
	// TODO: output logs in logstash json format
	// TODO: write output files as failures on termination
}
