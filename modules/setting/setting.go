package setting

import (
    "gopkg.in/ini.v1"
    "strings"
)

var (
    Cfg *ini.File

    host string
    port int
    database string

    query string
    headers []string
)

func init() {
    conf := "conf/app.ini"
    var err error
    Cfg, err = ini.Load(conf)
    if err != nil {
        panic(err)
    }

    err = loadCfg()
    if err != nil {
        panic(err)
    }
}

func loadCfg() (err error) {
    sec := Cfg.Section("impala")
    host = sec.Key("HOST").String()
    port, err = sec.Key("PORT").Int()
    if err != nil {
        return err
    }
    database = sec.Key("DATABASE").String()

    secRequest := Cfg.Section("request")
    query = secRequest.Key("QUERY").String()
    headers = strings.Split(secRequest.Key("HEADERS").String(), "|")
    return
}

func GetHost() string {
    return host
}

func GetPort() int {
    return port
}

func GetDatabase() string {
    return database
}

func GetQuery() string {
    return query
}

func GetHeaders() []string {
    return headers
}
