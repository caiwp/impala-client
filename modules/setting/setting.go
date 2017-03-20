package setting

import (
    "code.gitea.io/gitea/modules/log"
    "github.com/Unknwon/com"
    "gopkg.in/ini.v1"
    "os"
    "path"
    "strings"
    "fmt"
    "github.com/koblas/impalathing"
    "path/filepath"
)

var (
    Cfg *ini.File

    AppPath string
    CustomPath string
    CustomConf string

    // Log settings
    LogRootPath string
    LogModes    []string
    LogConfigs  []string
)

func init() {
    log.NewLogger(0, "console", `{"level": 0}`)

    var err error
    if AppPath, err = filepath.Abs("."); err != nil {
        log.Fatal(4, "failed to get app path: %v.", err)
    }
}

func WorkDir() string {
    return AppPath
}

func NewContext() {
    var err error
    Cfg = ini.Empty()

    workDir := WorkDir()

    conf := workDir + "/conf/app.ini"
    if com.IsFile(conf) {
        if err = Cfg.Append(conf); err != nil {
            log.Fatal(4, "failed to load conf '%s': %v.", conf, err)
        }
    }

    CustomPath = workDir + "/custom"
    CustomConf = CustomPath + "/conf/app.ini"

    if com.IsFile(CustomConf) {
        if err = Cfg.Append(CustomConf); err != nil {
            log.Fatal(4, "failed to load custom conf '%s': %v.", CustomConf, err)
        }
    } else {
        log.Warn("custom config '%s' not found.")
    }
    Cfg.NameMapper = ini.AllCapsUnderscore
    LogRootPath = Cfg.Section("").Key("ROOT_PATH").MustString(path.Join(workDir, "log"))
}

func NewServices() {
    newLogService()
    newImpalaService()
    newRequest()
}

var logLevels = map[string]string{
    "Trace": "0",
    "Debug": "1",
    "Info": "2",
    "Warn": "3",
    "Error": "4",
    "Critical": "5",
}

func newLogService() {
    LogModes = strings.Split(Cfg.Section("log").Key("MODE").MustString("console"), ",")
    LogConfigs = make([]string, len(LogModes))

    useConsole := false
    for i := 0; i < len(LogModes); i ++ {
        LogModes[i] = strings.TrimSpace(LogModes[i])
        if LogModes[i] == "console" {
            useConsole = true
        }
    }

    if !useConsole {
        log.DelLogger("console")
    }

    for i, mode := range LogModes {
        sec, err := Cfg.GetSection("log." + mode)
        if err != nil {
            sec, _ = Cfg.NewSection("log." + mode)
        }

        validLevels := []string{"Trace", "Debug", "Info", "Warn", "Error", "Critical"}
        // Log level.
        levelName := Cfg.Section("log." + mode).Key("LEVEL").In(
            Cfg.Section("log").Key("LEVEL").In("Trace", validLevels),
            validLevels)
        level, ok := logLevels[levelName]
        if !ok {
            log.Fatal(4, "Unknown log level: %s", levelName)
        }

        // Generate log configuration.
        switch mode {
        case "console":
            LogConfigs[i] = fmt.Sprintf(`{"level":%s}`, level)
        case "file":
            logPath := sec.Key("FILE_NAME").MustString(path.Join(LogRootPath, "impala-client.log"))
            if err = os.MkdirAll(path.Dir(logPath), os.ModePerm); err != nil {
                panic(err.Error())
            }

            LogConfigs[i] = fmt.Sprintf(
                `{"level":%s,"filename":"%s","rotate":%v,"maxlines":%d,"maxsize":%d,"daily":%v,"maxdays":%d}`, level,
                logPath,
                sec.Key("LOG_ROTATE").MustBool(true),
                sec.Key("MAX_LINES").MustInt(1000000),
                1 << uint(sec.Key("MAX_SIZE_SHIFT").MustInt(28)),
                sec.Key("DAILY_ROTATE").MustBool(true),
                sec.Key("MAX_DAYS").MustInt(7))
        }

        log.NewLogger(Cfg.Section("log").Key("BUFFER_LEN").MustInt64(10000), mode, LogConfigs[i])
        log.Info("Log Mode: %s(%s)", strings.Title(mode), levelName)
    }
}

type Impala struct {
    Host     string `ini:"HOST"`
    Port     int    `ini:"PORT"`
    Database string `ini:"DATABASE"`
}

var (
    ImplConn *impalathing.Connection
)

func newImpalaService() {
    impl := new(Impala)
    err := Cfg.Section("impala").MapTo(&impl)
    if err != nil {
        log.Fatal(4, "set impala config failed: %v", err)
    }

    ImplConn, err = impalathing.Connect(impl.Host, impl.Port, impalathing.DefaultOptions)
    if err != nil {
        log.Fatal(4, "connect impala failed [%s:%d]: %v", impl.Host, impl.Port, err)
    }

    _, err = ImplConn.Query(fmt.Sprintf("Use %s", impl.Database))
    if err != nil {
        log.Fatal(4, "select default database failed: %v", err)
    }

}

type Request struct {
    Query   string `ini:"QUERY"`
    Headers []string
}

var Req Request

func newRequest() {
    err := Cfg.Section("request").MapTo(&Req)
    if err != nil {
        log.Fatal(4, "set request config failed: %v", err)
    }

    if Req.Query == "" {
        log.Fatal(4, "request query empty")
    }

    Req.Headers = getHeaders(Req.Query)
}

const (
    SELECT = "select"
    FROM = "from"
    AS = ` as `
)

func getHeaders(q string) []string {
    var h []string
    q = strings.ToLower(q)
    q = q[len(SELECT):strings.Index(q, FROM)]

    sl := strings.Split(q, ",")
    for _, v := range sl {
        v = strings.TrimSpace(v)
        if strings.Contains(v, AS) {
            h = append(h, v[strings.LastIndex(v, AS) + len(AS):])
        } else {
            h = append(h, v[strings.LastIndex(v, ".") + len("."):])
        }
    }

    return h
}