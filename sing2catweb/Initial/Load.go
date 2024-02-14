package initial

import (
	"encoding/json"

	"github.com/bitly/go-simplejson"
	"github.com/spf13/viper"
)

func LoadTemplate(path string) {
	viper.SetConfigFile(path)
	viper.ReadInConfig()
	template,_ := json.Marshal(viper.AllSettings())
	template_json,_ := simplejson.NewJson(template)
	SetValue("template",template_json)
}
func LoadConfig(path string) {
	viper.SetConfigFile(path)
	viper.ReadInConfig()
	config,_ := json.Marshal(viper.AllSettings())
	config_json,_ := simplejson.NewJson(config)
	SetValue("config",config_json)
}
func GetBaseDir(){
	// base_dir := filepath.Dir(os.Args[0])
	base_dir := "E:/Myproject/sing2catweb"
	SetValue("base_dir",base_dir)
}