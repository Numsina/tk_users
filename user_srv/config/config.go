package config

type MysqlConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	UserName string `mapstructure:"username" json:"username"`
	PassWord string `mapstructure:"password" json:"password"`
	DBName   string `mapstructure:"dbname" json:"dbname"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	PassWord string `mapstructure:"password" json:"password"`
}

type JWTConfig struct {
	Key string `mapstructure:"key" json:"key"`
}

type ConsulConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Name     string `mapstructure:"name" json:"name"`
	Interval string `mapstructure:"interval" json:"interval"`
	Timeout  string `mapstructure:"timeout" json:"timeout"`
	Address  string `mapstructure:"address" json:"address"`
}

type NacosConfig struct {
	Host                string `mapstructure:"host" json:"host"`
	Port                uint64 `mapstructure:"port" json:"port"`
	NameSpaceId         string `mapstructure:"namespace_id" json:"namespace_id"`
	DataId              string `mapstructure:"data_id" json:"data_id"`
	Group               string `mapstructure:"group" json:"group"`
	TimeoutMs           uint64 `mapstructure:"timeout_ms" json:"timeout_ms"`
	NotLoadCacheAtStart bool   `mapstructure:"not_load_cache_at_start" json:"not_load_cache_at_start"`
	LogDir              string `mapstructure:"log_dir" json:"log_dir"`
	CacheDir            string `mapstructure:"cache_dir" json:"cache_dir"`
	LogLevel            string `mapstructure:"log_level" json:"log_level"`
}

type JaegerConfig struct {
	Host     string  `mapstructure:"host" json:"host"`
	Port     int     `mapstructure:"port" json:"port"`
	Name     string  `mapstructure:"name" json:"name"`
	Param    float64 `mapstructure:"param" json:"param"`
	LogSpans bool    `mapstructure:"log_spans" json:"log_spans"`
}

type Config struct {
	MysqlInfo   MysqlConfig  `mapstructure:"mysql" json:"mysql"`
	RedisInfo   RedisConfig  `mapstructure:"redis" json:"redis"`
	JwtInfo     JWTConfig    `mapstructure:"jwt" json:"jwt"`
	ConsuleInfo ConsulConfig `mapstructure:"consul" json:"consul"`
	NacosInfo   NacosConfig  `mapstructure:"nacos" json:"nacos"`
	JaegerInfo  JaegerConfig `mapstructure:"jaeger" json:"jaeger"`
}
