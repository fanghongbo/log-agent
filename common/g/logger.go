package g

import (
	"fmt"
	"github.com/fanghongbo/log-agent/utils"
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"io"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

var (
	AppLog *appLogger
)

func init() {
	AppLog = newAppLogger()
	log.SetFlags(log.Ldate | log.Ltime)
}

func initAppLog() error {
	var err error

	if err = AppLog.SetLogger(Cfg.config.Log, debug); err != nil {
		return err
	}

	return nil
}

type Logger struct {
	*log.Logger
}

func newLogger() *Logger {
	return &Logger{Logger: new(log.Logger)}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[INFO] %v", fmt.Sprintf(format, v...)))
}

func (l *Logger) Print(v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[INFO] %v", fmt.Sprint(v...)))
}

func (l *Logger) Println(v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[INFO] %v", fmt.Sprintln(v...)))
}

func (l *Logger) Info(v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[INFO] %v", fmt.Sprintln(v...)))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[INFO] %v", fmt.Sprintf(format, v...)))
}

func (l *Logger) Warn(v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[WARN] %v", fmt.Sprintln(v...)))
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[WARN] %v", fmt.Sprintf(format, v...)))
}

func (l *Logger) Error(v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[ERROR] %v", fmt.Sprintln(v...)))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[ERROR] %v", fmt.Sprintf(format, v...)))
}

func (l *Logger) Debug(v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[DEBUG] %v", fmt.Sprintln(v...)))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[DEBUG] %v", fmt.Sprintf(format, v...)))
}

func (l *Logger) Fatal(v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[FATAL] %v", fmt.Sprint(v...)))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[FATAL] %v", fmt.Sprintf(format, v...)))
	os.Exit(1)
}

func (l *Logger) Fatalln(v ...interface{}) {
	_ = l.Output(2, fmt.Sprintf("[FATAL] %v", fmt.Sprintln(v...)))
	os.Exit(1)
}

func (l *Logger) Panic(v ...interface{}) {
	s := fmt.Sprintf("[PANIC] %v", fmt.Sprint(v...))
	_ = l.Output(2, s)
	panic(s)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf("[PANIC] %v", fmt.Sprintf(format, v...))
	_ = l.Output(2, s)
	panic(s)
}

type appLogger struct {
	config *logConfig
	*sync.RWMutex
	fingerprint string
	logger      *Logger
}

func newAppLogger() *appLogger {
	return &appLogger{RWMutex: new(sync.RWMutex)}
}

func (u *appLogger) Logger() *log.Logger {
	u.RLock()
	defer u.RUnlock()

	if u.logger == nil {
		return new(log.Logger)
	}

	return u.logger.Logger
}

func (u *appLogger) Fingerprint() string {
	u.RLock()
	defer u.RUnlock()
	return u.fingerprint
}

func (u *appLogger) SetFingerprint(md5 string) {
	u.Lock()
	defer u.Unlock()
	u.fingerprint = md5
}

func (u *appLogger) SetLogger(config *logConfig, isDebug bool) error {
	var (
		rl      *rotateLogs.RotateLogs
		logFile string
		writers []io.Writer
		err     error
	)

	if path.IsAbs(config.Path) {
		logFile = path.Join(config.Path, "run.log")
	} else {
		if Pwd == "" {
			if Pwd, err = utils.GetPwd(); err != nil {
				return err
			}
		}

		logFile = path.Join(Pwd, path.Join(config.Path, "run.log"))
	}

	rl, err = rotateLogs.New(
		logFile+".%Y%m%d%H%M",
		rotateLogs.WithLinkName(logFile),
		rotateLogs.WithMaxAge(time.Duration(config.Rotate)*time.Hour),
		rotateLogs.WithRotationTime(time.Hour),
	)

	if err != nil {
		return fmt.Errorf("failed to initialize logger: %s", err)
	}

	writers = []io.Writer{
		rl,
	}

	// 同时输出到终端和日志文件
	if isDebug {
		writers = append(writers, os.Stdout)
	}

	u.Lock()

	u.logger = newLogger()
	u.logger.SetOutput(io.MultiWriter(writers...))
	u.logger.SetFlags(log.Ldate | log.Ltime)
	u.Unlock()

	return nil
}

func (u *appLogger) Info(v ...interface{}) {
	u.RLock()
	defer u.RUnlock()
	if u.logger != nil {
		u.logger.Info(v...)
	} else {
		log.Println("[INFO]", fmt.Sprintf("%v", v...))
	}
}

func (u *appLogger) Infof(format string, v ...interface{}) {
	u.RLock()
	defer u.RUnlock()
	if u.logger != nil {
		u.logger.Infof(format, v...)
	} else {
		log.Println("[INFO]", fmt.Sprintf(format, v...))
	}
}

func (u *appLogger) Warn(v ...interface{}) {
	u.RLock()
	defer u.RUnlock()
	if u.logger != nil {
		u.logger.Warn(v...)
	} else {
		log.Println("[WARN]", fmt.Sprintf("%v", v...))
	}
}

func (u *appLogger) Warnf(format string, v ...interface{}) {
	u.RLock()
	defer u.RUnlock()
	if u.logger != nil {
		u.logger.Warnf(format, v...)
	} else {
		log.Println("[WARN]", fmt.Sprintf(format, v...))
	}
}

func (u *appLogger) Error(v ...interface{}) {
	u.RLock()
	defer u.RUnlock()
	if u.logger != nil {
		u.logger.Error(v...)
	} else {
		log.Println("[ERROR]", fmt.Sprintf("%v", v...))
	}
}

func (u *appLogger) Errorf(format string, v ...interface{}) {
	u.RLock()
	defer u.RUnlock()
	if u.logger != nil {
		u.logger.Errorf(format, v...)
	} else {
		log.Println("[ERROR]", fmt.Sprintf(format, v...))
	}
}

func (u *appLogger) Debug(v ...interface{}) {
	u.RLock()
	defer u.RUnlock()
	if u.logger != nil {
		u.logger.Debug(v...)
	} else {
		log.Println("[DEBUG]", fmt.Sprintf("%v", v...))
	}
}

func (u *appLogger) Debugf(format string, v ...interface{}) {
	u.RLock()
	defer u.RUnlock()
	if u.logger != nil {
		u.logger.Debugf(format, v...)
	} else {
		log.Println("[DEBUG]", fmt.Sprintf(format, v...))
	}
}

func (u *appLogger) Fatal(v ...interface{}) {
	u.RLock()
	defer u.RUnlock()
	if u.logger != nil {
		u.logger.Fatal(v...)
	} else {
		log.Println("[FATAL]", fmt.Sprintf("%v", v...))
	}
	os.Exit(1)
}

func (u *appLogger) Fatalf(format string, v ...interface{}) {
	u.RLock()
	defer u.RUnlock()
	if u.logger != nil {
		u.logger.Fatalf(format, v...)
	} else {
		log.Println("[FATAL]", fmt.Sprintf(format, v...))
	}
	os.Exit(1)
}
