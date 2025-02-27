package initialize

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"

	"github.com/Numsina/tk_users/user_web/config"
)

var Conf *config.Config
var nacos = &config.NacosConfig{}

func InitConfig() *config.Config {
	if Conf == nil {
		v := viper.New()
		v.AutomaticEnv()
		configName := "config_dev.yaml"
		ok := v.GetBool("dev")
		if ok {
			configName = "config_pro.yaml"
		}
		dir, _ := os.Getwd()
		configName = path.Join(dir, configName)
		v.SetConfigFile(configName)
		v.SetConfigType("yaml")

		if err := v.ReadInConfig(); err != nil {
			panic(err)
		}

		log.Println("正在初始化用户服务nacos配置....")
		if err := v.Unmarshal(nacos); err != nil {
			log.Printf("用户服务nacos初始化配置失败...., %v", err)
			panic(err)
		}
		log.Println("用户服务nacos初始化成功....")

		// 初始化配置中心
		ccf := constant.ClientConfig{
			NamespaceId:         nacos.NameSpaceId,
			TimeoutMs:           nacos.TimeoutMs,
			NotLoadCacheAtStart: nacos.NotLoadCacheAtStart,
			LogDir:              nacos.LogDir,
			CacheDir:            nacos.CacheDir,
			LogLevel:            nacos.LogLevel,
		}

		scf := []constant.ServerConfig{
			{
				IpAddr: nacos.Host,
				Port:   nacos.Port,
			},
		}

		clientCfg, err := clients.NewConfigClient(
			vo.NacosClientParam{
				ClientConfig:  &ccf,
				ServerConfigs: scf,
			},
		)

		if err != nil {
			log.Printf("nacos配置客户端创建失败， %v", err)
			panic(err)
		}

		content, err := clientCfg.GetConfig(vo.ConfigParam{
			DataId: nacos.DataId,
			Group:  nacos.Group,
		})

		if err != nil {
			log.Printf("获取nacos配置中的内容失败， %v", err)
			panic(err)
		}

		err = json.Unmarshal([]byte(content), &Conf)
		if err != nil {
			log.Printf("json 反序列化失败， %v", err)
			panic(err)
		}

		return Conf
	}

	return Conf
}
