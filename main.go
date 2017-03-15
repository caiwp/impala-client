package main

import (
    "fmt"
    "time"

    "github.com/caiwp/impala-client/modules/setting"
    "github.com/koblas/impalathing"
    "github.com/caiwp/impala-client/modules/log"
    "github.com/caiwp/impala-client/modules/table"
)

var (
    l = log.NewLogger("default")

    impalaRes []map[string]interface{}
)

func main() {
    t0 := time.Now()
    defer func() {
        l.Info("End time duration: %.4fs", time.Since(t0).Seconds())
    }()

    host := setting.GetHost()
    port := setting.GetPort()
    database := setting.GetDatabase()

    conn, err := impalathing.Connect(host, port, impalathing.DefaultOptions)
    if err != nil {
        l.Error("connet failed: %s", err)
        return
    }
    defer conn.Close()

    _, err = conn.Query(fmt.Sprintf("USE %s", database))

    if err != nil {
        panic(err)
    }

    var query string
    query = setting.GetQuery()
    l.Warning(query)

    res, err := conn.Query(query)

    impalaRes = res.FetchAll()

    table.Show(impalaRes, setting.GetHeaders())
}
