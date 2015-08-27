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

package mysqltest

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"errors"
	"time"
)

var counts_to_check = []int64{1, 10, 50, 100, 150, 200, 250, 300, 600, 1200, 2400}
var seconds_to_check = []int64{0, 10, 30, 60, 120, 300, 600, 1200, 2400}

type MysqlTest struct {
	Name          string
	reportdir     string
	host          string
	user          string
	port          int
	pass          string
	interval      int
	timeout       int
	db            *sql.DB
	jsonlog       *os.File
	iteration     uint64
	maxjobs       uint
}


func RunMysqlTest(name string, host string, port int, user string, pass string, interval int, timeout int, reportdir string, jsonlog string) *MysqlTest {

	jsonlogfile, err := os.OpenFile(jsonlog, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Sprintf("Couldn't open jsonlog \"%s\" for writing: %s", jsonlog, err.Error()))
	}
	defer jsonlogfile.Close()

	m := MysqlTest{Name: name, host: host, port: port, user: user, pass: pass, interval: interval, timeout: timeout, reportdir: reportdir, jsonlog: jsonlogfile, iteration: 1}

	m.Run()

  return &m
}

func (t *MysqlTest) Run() {
	if t.interval < 0 {
		panic("interval must be a positive integer, or 0 to run the tests only once.")
	}
	t.RunOnceWithTimeout() // Run first test instantly.
	if t.interval > 0 {
		// run checks on intervals
		for _ = range time.Tick(time.Duration(t.interval) * time.Millisecond) {
			t.iteration++ // when this overflows it will become 0 with no problems http://play.golang.org/p/fbjwHKcIaU
			t.RunOnceWithTimeout()
		}
	}
}

// This is a very dumb json func. If more interesting stuff needs to be logged,
// pass it in as a map[string]interface{} and then detect value as int, string, w/e
// before marshaling json.
func (t *MysqlTest) JsonLog(msg string) {
	t.jsonlog.WriteString(fmt.Sprintf("{\"@timestamp\":\"%s\",\"type\":\"mysql_probe\",\"host\":\"%s\",\"iteration\":%v,\"message\":\"%s\"}\n",
		time.Now().Format(time.RFC3339), t.host, t.iteration, msg))
}

func (t *MysqlTest) GetWeight(val int64, max int64) string {
	if val == 0 {
		return "100%"
	}
	return fmt.Sprintf("%d%%", 100-(100*(val/max)))
}

func (t *MysqlTest) RunOnceWithTimeout() {
  timeout_ch := make(chan *MysqlTestResult, 1)
  ch := make(chan *MysqlTestResult, 1)

  // after 1 second, fill a blank response and
  // send that indicates everything timed out
  go func() {
    time.Sleep(time.Duration(t.timeout) * time.Millisecond)
    // fill timeout statuses up
    res := MysqlTestResult{}
    res.AddTextResult("connect", "down # timeout")
    for _,c := range counts_to_check {
      res.AddTextResult(fmt.Sprintf("connection_count_lte_%d", c), "down # timeout")
    }
    for _,c := range seconds_to_check {
       res.AddTextResult(fmt.Sprintf("replication_delay_lte_%d", c), "down # timeout")
    }
    timeout_ch <- &res
  }()

  // run the actual check with no delay
  // hopefully it wins.
  go func() {
    res := t.RunOnce()
    ch <- res
  }()

  var res * MysqlTestResult
  // grab the first thing that comes back
  select {
  case res = <-ch:
  case res = <-timeout_ch:
  }

  writer := MysqlTestResultWriter{test: t, result: res}
  writer.WriteResult()
}

func (t *MysqlTest) RunOnce() *MysqlTestResult {
        res := MysqlTestResult{}
	start := time.Now()
	err := t.Connect()
	dur := time.Since(start)
	if err != nil {
		res.AddTextResult("connect", fmt.Sprintf("down # %s", err.Error()))
	} else {
		res.AddTextResult("connect", fmt.Sprintf("up 100%% # connect took %s", dur.String()))
	}

	start = time.Now()
	count, err := t.CountConnections()
	if err != nil {
		description := fmt.Sprintf("Connection count check failed: %s", err.Error())
		for _, connections := range counts_to_check {
			res.AddTextResult(fmt.Sprintf("connection_count_lte_%d", connections), fmt.Sprintf("down # Connection count test (%d connections <= %d)? : %s", count, connections, description))
		}
	} else {
		description := fmt.Sprintf("Connection count is %d", count)
		for _, connections := range counts_to_check {
			var status string
			if count <= connections {
				status = fmt.Sprintf("up %s", t.GetWeight(count, connections))
			} else {
				status = "down"
			}
			res.AddTextResult(fmt.Sprintf("connection_count_lte_%d", connections), fmt.Sprintf("%s # Connection count test (%d connections <= %d)? : %s", status, count, connections, description))
		}
	}

	delay, description, err := t.CheckReplication()
	if err != nil {
		description = fmt.Sprintf("%s: %s", description, err.Error())
		for _, seconds := range seconds_to_check {
			res.AddTextResult(fmt.Sprintf("replication_delay_lte_%d", seconds), fmt.Sprintf("down # Replication delay test (%d delay <= %d seconds)? : %s", delay, seconds, description))
		}
	} else {
		for _, seconds := range seconds_to_check {
			var status string
			if delay <= seconds {
				status = fmt.Sprintf("up %s", t.GetWeight(delay, seconds))
			} else {
				status = "down"
			}
			res.AddTextResult(fmt.Sprintf("replication_delay_lte_%d", seconds), fmt.Sprintf("%s # Replication delay test (%d delay <= %d seconds)? : %s", status, delay, seconds, description))
		}
	}
	defer t.Disconnect()

  return &res
}

func (t *MysqlTest) Disconnect() {
	//fmt.Println("Disconnecting")
	if t.db != nil {
		t.db.Close()
	}
}

// NOTE: only works if one master
func (t *MysqlTest) CheckReplication() (int64, string, error) {
	var description string
	if t.db == nil {
		description = "DB connection is invalid"
		return 0, description, errors.New(description)
	}
	rows, err := t.db.Query("SHOW SLAVE STATUS") // queryable from PERFORMANCE_SCHEMA at mysql 5.7.2: http://bugs.mysql.com/bug.php?id=35994
	if err != nil {
		description = "Query 'SHOW SLAVE STATUS' failed"
		t.JsonLog(fmt.Sprintf("%s: %s", description, err.Error()))
		t.db = nil
		return 0, description, err
	}

	// Since our mysql is too old to use PERFORMANCE_SCHEMA and select only the columns we want,
	// we need to find the column count and use RawBytes to be a placeholder for all the columns we don't care about.
	// http://go-database-sql.org/varcols.html
	cols, err := rows.Columns() // Remember to check err afterwards
	if err != nil {
		description = "Couldn't retrieve column information for 'SHOW SLAVE STATUS' statement"
		t.JsonLog(fmt.Sprintf("%s: %s", description, err.Error()))
		t.db = nil
		return 0, description, err
	}
	vals := make([]interface{}, len(cols))
	var slave_io_running string
	var seconds_behind_master int64
	seconds_behind_master = -1
	for i, _ := range cols {
		switch cols[i] {
		case "Slave_IO_Running":
			vals[i] = &slave_io_running
		case "Seconds_Behind_Master":
			vals[i] = &seconds_behind_master
		default:
			vals[i] = new(sql.RawBytes)
		}
	}
	err = nil
	for rows.Next() {
		err = rows.Scan(vals...)
		// Now you can check each element of vals for nil-ness,
		// and you can use type introspection and type assertions
		// to fetch the column into a typed variable.
	}
	if slave_io_running == "" || seconds_behind_master == -1 {
		description = fmt.Sprintf("Slave status and/or seconds_behind_master could not be determined. slave_io_running: %s, seconds_behind_master: %d", slave_io_running, seconds_behind_master)
		err = errors.New(description)
	} else if slave_io_running == "No" {
		description = "Slave is not running"
		err = errors.New(description)
	} else {
		description = fmt.Sprintf("Slave running: %s Seconds behind: %d", slave_io_running, seconds_behind_master)
	}

	t.JsonLog(description)
	return seconds_behind_master, description, err
}

func (t *MysqlTest) CountConnections() (int64, error) {
	if t.db == nil {
		return -2, nil
	}
	row := t.db.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.PROCESSLIST WHERE USER != 'system user'")
	var processcount int64
	err := row.Scan(&processcount)
	if err != nil {
		t.JsonLog(fmt.Sprintf("Couldn't determine connection count: %s", err.Error()))
		return -1, err
	}
	processcount = processcount - 1 // Deduct this connection from teh count.
	t.JsonLog(fmt.Sprintf("Process count: %d", processcount))
	return processcount, nil
}

func (t *MysqlTest) Connect() error {
	t.JsonLog(fmt.Sprintf("Connecting to %s@%s:%d", t.user, t.host, t.port))

	// Create dsn like such https://github.com/Go-SQL-Driver/MySQL/#dsn-data-source-name
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	// username:password@protocol(address)/dbname?param=value
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=%dms&allowOldPasswords=1", t.user, t.pass, t.host, t.port, t.timeout)
	//fmt.Println("DSN: ", dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.JsonLog(fmt.Sprintf("Couldn't open mysql connection: %s", err.Error()))
		t.db = nil
		return err
	}

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		t.JsonLog(fmt.Sprintf("Couldn't connect to mysql server: %s", err.Error()))
		t.db = nil
		return err
	}

	//fmt.Println("SUCCESS")
	t.db = db
	db.Query("SET SESSION wait_timeout=1")
	return nil
}
