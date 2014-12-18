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
	"log"
	"os"
	"path/filepath"
	"time"
)

type MysqlTest struct {
	Name          string
	filedirectory string
	host          string
	user          string
	port          int
	pass          string
	interval      int
	timeout       int
	db            *sql.DB
}

func NewMysqlTest(name string, host string, port int, user string, pass string, interval int, timeout int, filedirectory string) *MysqlTest {
	m := MysqlTest{Name: name, host: host, port: port, user: user, pass: pass, interval: interval, timeout: timeout, filedirectory: filedirectory}
	return &m
}

func (t *MysqlTest) Run() {
	if t.interval <= 0 {
		panic("interval must be a positive integer")
	}
	// run checks on intervals
	for _ = range time.Tick(time.Duration(t.interval) * time.Millisecond) {
		// TODO: expire a test with a timeout
		t.RunOnce()
	}
}

// func writeHttpResult(filedirectory string) {
//
// }
func (t *MysqlTest) RunOnce() {
	start := time.Now()
	err := t.Connect()
	dur := time.Since(start)
	if err != nil {
		t.WriteResult("connect", false, err.Error())
		return
	}
	t.WriteResult("connect", true, fmt.Sprintf("connect took %s", dur.String()))

	start = time.Now()
	count, err := t.CountConnections()
	if err != nil {
		t.WriteResult("connection_count", false, err.Error())
		return
	}
	result := fmt.Sprintf("Process count %d", count)
	// TODO: check connection count and pass only if below threshold.
	t.WriteResult("connection_count", true, result)
	fmt.Println(result)

	_, _, err = t.CheckReplication()
	defer t.Disconnect()
}

func (t *MysqlTest) Disconnect() {
	fmt.Println("Disconnecting")
	if t.db != nil {
		t.db.Close()
	}
}

func (t *MysqlTest) WriteResult(testname string, passed bool, description string) {
	status := "500 Internal Server Error"
	if passed {
		status = "200 OK"
	}
	now := time.Now().Format(time.RFC1123Z)
	response := fmt.Sprintf("HTTP/1.1 %s\r\nDate: %s\r\nContent-Type: text/plain\r\n\r\n%s\r\n", status, now, description)

	file, err := os.OpenFile(fmt.Sprintf("%s.txt", filepath.Join(t.filedirectory, "/", testname)),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		//log.Fatal(err)
		log.Print(err.Error())
	}
	defer file.Close()

	file.WriteString(response)
}

func (t *MysqlTest) CheckReplication() (int64, string, error) {
	rows, err := t.db.Query("SHOW SLAVE STATUS") // queryable from PERFORMANCE_SCHEMA at mysql 5.7.2: http://bugs.mysql.com/bug.php?id=35994
	if err != nil {
		return 0, "Query 'SHOW SLAVE STATUS' failed", err
	}

	// Since our mysql is too old to use PERFORMANCE_SCHEMA and select only the columns we want,
	// we need to find the column count and use RawBytes to be a placeholder for all the columns we don't care about.
	// http://go-database-sql.org/varcols.html
	cols, err := rows.Columns() // Remember to check err afterwards
	if err != nil {
		return 0, "Couldn't retrieve column information for 'SHOW SLAVE STATUS' statement.", err
	}
	vals := make([]interface{}, len(cols))
	var slave_io_running string
	var seconds_behind_master int64
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
	for rows.Next() {
		err = rows.Scan(vals...)
		// Now you can check each element of vals for nil-ness,
		// and you can use type introspection and type assertions
		// to fetch the column into a typed variable.
		fmt.Println(fmt.Sprintf("Slave running: %s\nSeconds behind: %d", slave_io_running, seconds_behind_master))
	}

	return seconds_behind_master, "", nil
}

func (t *MysqlTest) CountConnections() (int64, error) {
	row := t.db.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.PROCESSLIST WHERE USER != 'system user' AND USER != 'mysql_probe'")
	var processcount int64
	err := row.Scan(&processcount)
	if err != nil {
		return -1, err
	}
	return processcount, nil
}

func (t *MysqlTest) Connect() error {
	fmt.Println("Host: ", t.host, " Port: ", t.port, " User: ", t.user, " Pass: ", t.pass)

	// Create dsn like such https://github.com/Go-SQL-Driver/MySQL/#dsn-data-source-name
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	// username:password@protocol(address)/dbname?param=value
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=%dms&allowOldPasswords=1", t.user, t.pass, t.host, t.port, t.timeout)
	fmt.Println("DSN: ", dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		return err
	}

	fmt.Println("SUCCESS")
	t.db = db
	return nil
}
