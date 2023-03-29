package bot

import "github.com/spf13/viper"

// 机器人配置类
type Config struct {
	FeishuAppId                string
	FeishuAppSecret            string
	FeishuAppEncryptKey        string
	FeishuAppVerificationToken string
	OpenAIKeys                 []string
	OpenAIUrl                  string
	HttpPort                   int
}

// 读取配置文件
func LoadConfig(path string) *Config {
	viper.SetConfigFile(path)
	viper.ReadInConfig()
	viper.AutomaticEnv()

	config := &Config{
		FeishuAppId:                viper.GetString("APP_ID"),
		FeishuAppSecret:            viper.GetString("APP_SECRET"),
		FeishuAppEncryptKey:        viper.GetString("APP_ENCRYPT_KEY"),
		FeishuAppVerificationToken: viper.GetString("APP_VERIFICATION_TOKEN"),
		OpenAIKeys:                 viper.GetStringSlice("OPENAI_KEYS"),
		OpenAIUrl:                  getViperStr("OPENAI_URL", "https://api.openai.com"),
		HttpPort:                   getViperInt("HTTP_PORT", 10086),
	}

	// 无效的配置直接panic
	if !config.Valid() {
		panic("invalid bot config")
	}

	return config
}

func getViperInt(key string, defaultValue int) int {
	if viper.IsSet(key) {
		return viper.GetInt(key)
	}
	return defaultValue
}

func getViperStr(key string, defaultValue string) string {
	if viper.IsSet(key) {
		return viper.GetString(key)
	}
	return defaultValue
}

func (c *Config) Valid() bool {
	return c.FeishuAppId != "" &&
		c.FeishuAppSecret != "" &&
		c.FeishuAppVerificationToken != "" &&
		len(c.OpenAIKeys) != 0
}
