package sing2catconfig

import (
	"errors"
	"fmt"

	"github.com/huandu/go-clone"
	"github.com/spf13/viper"
)

var Global_vars = make(map[string]interface{})

func Get_value(key string) (interface{}, error) {
	result := clone.Clone(Global_vars[key])
	// 判断result是否为空,为空则报错
	if result != nil {
		return result, nil
	} else {
		err := fmt.Sprintf("Key %s in Global_vars not found",key)
		return result,errors.New(err)
	}
}
func Set_value(key string, value interface{}) {
	Global_vars[key] = value
}

func Get_Sing2cat_dir() string {
	// base_dir := filepath.Dir(os.Args[0])
	base_dir := "E:/Myproject/sing2cat_web"
	return base_dir
}
func get_sing2cat_config(file string){
	// 获取项目目录路径,获取失败直接panic退出该进程
	project_dir,err := Get_value("project_dir")
	if err != nil{
		Logger_caller(fmt.Sprintf("Get %s Dir failed!",file),err)
		panic(fmt.Sprintf("Get %s Dir failed!",file))
	}
	// 读取配置文件,读取错误则panic退出该进程
	viper.SetConfigFile(fmt.Sprintf("%s/config/sing2cat/%s.yaml",project_dir,file))
	err = viper.ReadInConfig()
	if err != nil{
		Logger_caller(fmt.Sprintf("Read %s failed!",file),err)
		panic(fmt.Sprintf("Read %s failed!",file))
	}
	Set_value(file,viper.AllSettings())
}

func Sing2cat_Init(){

	get_sing2cat_config("config")
	get_sing2cat_config("template")
	// 日志记录
	Logger_caller("Initial completed!",nil)
}