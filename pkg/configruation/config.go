package configruation

import (
	"carina/utils"
	"carina/utils/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"os"
	"strings"
)

// 配置文件路径
const (
	configPath        = "/etc/carina/"
	SchedulerBinpack  = "binpack"
	SchedulerSpradout = "spradout"
	diskGroupType     = "type"
)

// 提供给其他应用获取服务数据
// 这个configMap理论上应该由Node Server更新，为了实现简单改为有Control Server更新，遍历所有Node信息更新configmap
// 暂定这些参数字段，不排除会增加一些需要暴露的数据
type ConfigProvider struct {
	NodeIp string   `json:"nodeip"`
	Vg     []string `json:"vg"`
}

var GlobalConfig2 *viper.Viper

func init() {
	log.Info("Loading global configuration ...")
	GlobalConfig2 = initConfig()
	go dynamicConfig()

}

func initConfig() *viper.Viper {
	GlobalConfig := viper.New()
	GlobalConfig.AddConfigPath(configPath)
	GlobalConfig.SetConfigName("config")
	GlobalConfig.SetConfigType("json")
	err := GlobalConfig.ReadInConfig()
	if err != nil {
		log.Error("Failed to get the configuration")
		os.Exit(-1)
	}
	return GlobalConfig
}

func dynamicConfig() {
	GlobalConfig2.WatchConfig()
	GlobalConfig2.OnConfigChange(func(event fsnotify.Event) {
		log.Info("Detect config change: %s", event.String())
	})
}

// 支持正则表达式
// 定时扫描本地磁盘，凡是匹配的将被加入到相应vg卷组
// 对于此配置的修改需要非常慎重，如果更改匹配条件，可能会移除正在使用的磁盘
func DiskSelector() []string {
	diskSelector := GlobalConfig2.GetStringSlice("diskSelector")
	if len(diskSelector) == 0 {
		log.Warn("No device is initialized because there is no configuration")
	}
	return diskSelector
}

// 定时磁盘扫描时间间隔(秒),默认60s
func DiskScanInterval() int64 {
	diskScanInterval := GlobalConfig2.GetInt64("diskScanInterval")
	if diskScanInterval < 300 {
		diskScanInterval = 300
	}
	return diskScanInterval
}

// 磁盘分组策略，目前只支持根据磁盘类型分组
func DiskGroupPolicy() string {
	diskGroupPolicy := GlobalConfig2.GetString("diskGroupPolicy")
	diskGroupPolicy = "type"
	return diskGroupPolicy

}

// pv调度策略binpac/spradout，默认为binpac
func SchedulerStrategy() string {
	schedulerStrategy := GlobalConfig2.GetString("schedulerStrategy")
	if utils.IsContainsString([]string{SchedulerBinpack, SchedulerSpradout}, strings.ToLower(schedulerStrategy)) {
		schedulerStrategy = strings.ToLower(schedulerStrategy)
	} else {
		schedulerStrategy = SchedulerBinpack
	}
	return schedulerStrategy
}
