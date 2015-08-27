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
	"github.com/codegangsta/cli"
	"github.com/haikulearning/mysql_probe/mysqltest"
	"github.com/haikulearning/mysql_probe/statusserver"
	_ "github.com/go-sql-driver/mysql"
	"os"
  "log"
  "sync"
)

// The current version of the app
const VERSION string = "0.0.3"

func test_mysql(c *cli.Context) {
  log.Println("Testing mysql server")
  // Run our mysql tests
  mysqltest.RunMysqlTest("connection", c.String("host"), c.Int("port"), c.String("user"), c.String("pass"), c.Int("interval"), c.Int("timeout"), c.String("reports"), c.String("jsonlog"))
}

func serve_status(c *cli.Context) {
  log.Println("Running status server")
  // Run our status server
  statusserver.StartStatuServer(c.String("reports"), c.Int("server_port"))
}

func test_and_serve(c *cli.Context) {
  var wg sync.WaitGroup

  wg.Add(1)
  go func() {
    // Decrement the counter when the goroutine completes.
    defer wg.Done()
    test_mysql(c)
  }()

  wg.Add(1)
  go func() {
    // Decrement the counter when the goroutine completes.
    defer wg.Done()
    serve_status(c)
  }()

  wg.Wait()
}

func main() {
	app := cli.NewApp()
	app.Name = "mysql_probe"
	app.Usage = "test mysql health, write the results to disk, and serve them via http"
	app.Version = VERSION

  test_mysql_flags := []cli.Flag{
		cli.StringFlag{
			Name:   "host",
			Value:  "127.0.0.1",
			Usage:  "(test) mysql host to connect to",
			EnvVar: "MYSQL_PROBE_HOST",
		},
		cli.IntFlag{
			Name:   "port, p",
			Value:  3306,
			Usage:  "(test) mysql port to connect to",
			EnvVar: "MYSQL_PROBE_PORT",
		},
		cli.StringFlag{
			Name:   "user, u",
			Value:  "root",
			Usage:  "(test) mysql username to connect with",
			EnvVar: "MYSQL_PROBE_USER",
		},
		cli.StringFlag{
			Name:   "pass",
			Value:  "test",
			Usage:  "(test) mysql password to connect with",
			EnvVar: "MYSQL_PROBE_PASS",
		},
		cli.IntFlag{
			Name:   "timeout, t",
			Value:  2000,
			Usage:  "(test) time in milliseconds to wait for a mysql connection",
			EnvVar: "MYSQL_PROBE_TIMEOUT",
		},
		cli.IntFlag{
			Name:   "interval, i",
			Value:  250,
			Usage:  "(test) interval in milliseconds to run the checks, set to 0 to only run the tests once",
			EnvVar: "MYSQL_PROBE_INTERVAL",
		},
		cli.StringFlag{
			Name:   "jsonlog",
			Value:  "/dev/stdout",
			Usage:  "(test) file to write test results log output in json",
			EnvVar: "MYSQL_PROBE_JSONLOG",
		},
  }
  both_flags := []cli.Flag{
		cli.StringFlag{
			Name:   "reports",
			Value:  "tmp",
			Usage:  "(all) directory for test results up/down status files",
			EnvVar: "MYSQL_PROBE_REPORTS",
		},
  }
  server_status_flags := []cli.Flag{
		cli.IntFlag{
			Name:   "server_port, s",
			Value:  3001,
			Usage:  "(serve) port number where the status server accepts requests",
			EnvVar: "MYSQL_PROBE_SERVER",
		},
  }

  all_flags := []cli.Flag{}
  all_flags = append(all_flags, test_mysql_flags...)
  all_flags = append(all_flags, both_flags...)
  all_flags = append(all_flags, server_status_flags...)

  app.Commands = []cli.Command{
    {
      Name: "start",
      Usage: "both test mysql & run a status server of the results",
      Action: func(c *cli.Context) {
        test_and_serve(c)
      },
      Flags: all_flags,
    },
    {
      Name: "test",
      Usage: "test a mysql server's status",
      Action: func(c *cli.Context) {
        test_mysql(c)
      },
      Flags: append(test_mysql_flags, both_flags...),
    },
    {
      Name: "serve",
      Usage: "run server of the test's output files",
      Action: func(c *cli.Context) {
        serve_status(c)
      },
      Flags: append(server_status_flags, both_flags...),
    },
  }
	app.EnableBashCompletion = true
	app.Run(os.Args)

	// TODO: write output files as failures on termination
}
