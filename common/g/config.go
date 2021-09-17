package g

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/fanghongbo/log-agent/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	cfgFile string
	Cfg     *cfgConfig
	debug   bool
)

func init() {
	Cfg = newCfgConfig()
}

type globalConfig struct {
	Hostname   string      `json:"hostname"`
	Endpoint   string      `json:"endpoint"`
	AccessKey  string      `json:"accessKey"`
	SecretKey  string      `json:"secretKey"`
	BucketName string      `json:"bucketName"`
	PartSize   int64       `json:"partSize"`
	Log        *logConfig  `json:"log"`
	Strategies []*strategy `json:"strategies"`
}

type strategy struct {
	Src         string `json:"src"`
	Dest        string `json:"dest"`
	Pattern     string `json:"pattern"`
	Exclude     string `json:"exclude"`
	ForbidWrite bool   `json:"forbidWrite"`
	AfterDelete bool   `json:"afterDelete"`
	BeforeTime  int64  `json:"beforeTime"`
	Sign        string `json:"sign"`
	Spec        string `json:"spec"`
}

type logConfig struct {
	Path   string `json:"path"`
	Rotate int    `json:"rotate"`
}

type cfgConfig struct {
	config      *globalConfig
	fingerprint string
	*sync.RWMutex
}

func newCfgConfig() *cfgConfig {
	return &cfgConfig{RWMutex: new(sync.RWMutex)}
}

func (u *cfgConfig) Config() *globalConfig {
	u.RLock()
	defer u.RUnlock()
	return u.config
}

func (u *cfgConfig) SetConfig(config *globalConfig) {
	u.Lock()
	defer u.Unlock()
	u.config = config
}

func (u *cfgConfig) Fingerprint() string {
	u.RLock()
	defer u.RUnlock()
	return u.fingerprint
}

func (u *cfgConfig) SetFingerprint(md5 string) {
	u.Lock()
	defer u.Unlock()
	u.fingerprint = md5
}

func (u *globalConfig) ParseVariable(str string) string {
	for _, item := range []string{"${hostname}", "${HOSTNAME}"} {
		str = strings.ReplaceAll(str, item, u.Hostname)
	}
	return str
}

func (u *globalConfig) Validator() error {
	var err error

	if u.Hostname == "" {
		u.Hostname = LocalIp
	}

	if u.Hostname == "" {
		return fmt.Errorf("hostname can not be empty")
	}

	if u.Endpoint == "" {
		return fmt.Errorf("endpoint can not be empty")
	}

	if u.AccessKey == "" {
		return fmt.Errorf("access key can not be empty")
	}

	if u.SecretKey == "" {
		return fmt.Errorf("secret key can not be empty")
	}

	if _, err = oss.New(u.Endpoint, u.AccessKey, u.SecretKey); err != nil {
		return fmt.Errorf("failed to verify alicloud ram account: %v", err.Error())
	}

	if u.BucketName == "" {
		return fmt.Errorf("bucket name can not be empty")
	}

	if u.PartSize <= 0 {
		u.PartSize = 100 * 1024
	}

	for _, item := range u.Strategies {
		var (
			exist bool
			err   error
		)

		if item.Src == "" {
			return fmt.Errorf("src path can not be empty")
		} else {
			if exist, err = utils.CheckPath(item.Src); err != nil {
				return err
			}

			if !exist {
				return fmt.Errorf("src path %s is not exist", item.Src)
			}

			item.Src = u.ParseVariable(item.Src)

			if !strings.HasSuffix(item.Src, string(filepath.Separator)) {
				item.Src = item.Src + string(filepath.Separator)
			}
		}

		if item.Dest == "" {
			return fmt.Errorf("dest can not be empty")
		} else {
			item.Dest = u.ParseVariable(item.Dest)

			if !strings.HasSuffix(item.Dest, string(filepath.Separator)) {
				item.Dest = item.Dest + string(filepath.Separator)
			}
		}

		if item.Pattern != "" {
			if !utils.CheckRegExpr(item.Pattern) {
				return fmt.Errorf("pattern regular is forbid")
			}
		}

		if item.Exclude != "" {
			if !utils.CheckRegExpr(item.Exclude) {
				return fmt.Errorf("exclude regular is forbid")
			}
		}

		if item.BeforeTime < 0 {
			item.BeforeTime = 0
		}

		if item.Spec == "" {
			return fmt.Errorf("cron spec can not be empty")
		}

		if !utils.CheckCronSpec(item.Spec) {
			fmt.Println("cron spec is forbid")
		}
	}

	if u.Log == nil {
		return fmt.Errorf("log config can not be empty")
	}

	if u.Log.Path == "" {
		return fmt.Errorf("log path can not be empty")
	}

	if u.Log.Rotate < 0 {
		u.Log.Rotate = 0
	}

	return nil
}

func initConfig() error {
	var (
		bs     []byte
		config *globalConfig
		err    error
	)

	if cfgFile == "" {
		return errors.New("config file not specified: use --config $filename")
	}

	if bs, err = ioutil.ReadFile(cfgFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file %s specified not found", cfgFile)
		} else {
			return fmt.Errorf("read config file failed: %s", err.Error())
		}
	} else {
		AppLog.Infof("use config file: %s", cfgFile)

		if err = json.Unmarshal(bs, &config); err != nil {
			return fmt.Errorf("decode cfg config file err: %s", err.Error())
		} else {
			AppLog.Infof("load cfg config success from %s", cfgFile)
		}
	}

	// 配置校验
	if err = config.Validator(); err != nil {
		return err
	}

	Cfg.SetConfig(config)
	Cfg.SetFingerprint(utils.GetStringMd5(string(bs)))

	return nil
}
