package cfg

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	L "team_streams/internal/log"
	T "team_streams/internal/types"
)

var _ T.ICfg = (*CfgMaps)(nil)

type CfgMaps struct {
	envFname  string
	envVals   map[string]string
	jsonFname string
	jsonVals  map[string]string
	log       T.ILog
}

func NewCfgMaps(dir, file string) *CfgMaps {
	envVals := make(map[string]string, 8)
	jsonVals := make(map[string]string, 8)
	envVals[T.TS_APP_NAME] = file
	envVals[T.TS_APP_IP] = "localhost"
	envVals[T.TS_LOG_LEVEL] = "INFO" // LOG levels: TRACE, DEBUG, INFO, WARN, ERROR, PANIC, FATAL, NOLOG(default if empty or mess)
	envVals[T.TG_BOT_TOKEN] = ""
	envVals[T.TTV_CLIENT_ID] = ""
	envVals[T.TTV_CLIENT_SECRET] = ""
	envVals[T.TTV_APPACCESS_TOKEN] = ""
	return &CfgMaps{
		envFname:  filepath.Join(dir, file+".env"),
		envVals:   envVals,
		jsonFname: filepath.Join(dir, file+".json"),
		jsonVals:  jsonVals,
	}
}

func (c *CfgMaps) Parse() T.ICfg {
	c.log = L.NewLogFprintf(c, 0, 0)
	c.parseIpFromInterface()
	if len(c.envFname) != 0 {
		if _, err := os.Stat(c.envFname); err == nil {
			c.parseFileDotEnvVars()
		}
	}
	c.parseOsEnvVars()
	return c
}

func (c *CfgMaps) SetEnvVal(setkey string, setval string) {
	for key := range c.envVals {
		if key == setkey {
			c.envVals[setkey] = setval
			envFileBuf, err := os.ReadFile(c.envFname)
			if err != nil {
				c.log.LogWarn("%s", "(CfgMaps).SetEnvVal(): warning while reading cfg file"+err.Error())
			}

			regex := regexp.MustCompile(setkey + "=" + ".*\n")
			fixedEnvFileBuf := regex.ReplaceAllString(string(envFileBuf), setkey+"="+setval+"\n")

			err = os.WriteFile(c.envFname, []byte(fixedEnvFileBuf), os.ModePerm)
			if err != nil {
				c.log.LogWarn("%s", "(CfgMaps).SetEnvVal(): warning while writing cfg file"+err.Error())
			}
		}
	}
}

func (c *CfgMaps) GetEnvVal(key string) string {
	val := c.envVals[key]
	return val
}

func (c *CfgMaps) parseOsEnvVars() {
	for key := range c.envVals {
		if v, ok := os.LookupEnv(key); ok && (len(v) > 0) {
			c.envVals[key] = v
			c.log.LogDebug("OSENV %s=%s\n", key, v)
		}
	}
}

func (c *CfgMaps) parseFileDotEnvVars() {
	f, err := os.Open(c.envFname)
	if err != nil {
		c.log.LogWarn("%s", "(CfgMaps).parseFileDotEnvVars(): warning while opening cfg file"+err.Error())
		return
	}
	c.log.LogDebug("load config from file: %s", c.envFname)
	defer f.Close()

	pattern := regexp.MustCompile("^[0-9A-Za-z_]+=[0-9A-Za-z_:/.]+")
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		str := pattern.FindString(scanner.Text())
		if len(str) > 0 {
			strarr := strings.Split(str, "=")
			if _, ok := c.envVals[strarr[0]]; ok {
				c.envVals[strarr[0]] = strarr[1]
				c.log.LogDebug("CFGFILE %s=%s\n", strarr[0], strarr[1])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		c.log.LogWarn("%s", "(CfgMaps).parseFileDotEnvVars(): warning while reading cfg file"+err.Error())
	}
}

func (c *CfgMaps) parseIpFromInterface() {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		c.log.LogWarn("%s", "CfgMaps.parseIpFromInterface(): warning while getting IP interface"+err.Error())
		return
	}
	strarr := []string{}
	for _, addr := range addr {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				strarr = append(strarr, ipnet.IP.String())
			}
		}
	}
	ip := strings.Join(strarr, ";")
	if len(ip) > 0 {
		c.SetEnvVal(T.TS_APP_IP, ip)
		c.log.LogDebug("(CfgMaps).parseIpFromInterface(): %s", ip)
	}
}
