package main

import (
    "time"

    "code.gitea.io/gitea/modules/log"
    "github.com/caiwp/impala-client/modules/setting"
    "github.com/caiwp/impala-client/modules/table"
)

func main() {
    GlobalInit()

    t0 := time.Now()
    defer func() {
        log.Warn("End time duration: %.4fs", time.Since(t0).Seconds())
        setting.ImplConn.Close()
        log.Close()
    }()

    query := setting.Req.Query
    log.Trace(query)

    res, err := setting.ImplConn.Query(query)
    if err != nil {
        log.Fatal(4, "request query [%s] failed: %v", query, err)
        return
    }

    impalaRes := res.FetchAll()

    table.Show(impalaRes, setting.Req.Headers)
    return
}

func GlobalInit() {
    setting.NewContext()
    setting.NewServices()
}