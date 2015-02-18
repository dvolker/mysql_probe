package mysqltest

type MysqlTestResult struct {
    Entries map[string]string
}

func (r * MysqlTestResult) AddTextResult(subject string, content string) {
    if r.Entries == nil {
      r.Entries = make(map[string]string)
    }
    r.Entries[subject] = content
}
