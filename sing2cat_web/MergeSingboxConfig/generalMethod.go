package mergesingboxconfig

import (
	"errors"
	"fmt"
	"os"
	sing2catconfig "sing2cat_web/Sing2catConfig"
)
func Write_config_file(config_file []byte)error{
	project_dir,err := sing2catconfig.Get_value("project_dir")
	if err != nil{
		sing2catconfig.Logger_caller("get project dir failed!",err)
		return err
	}
	// 写入文件
	dst_dir := fmt.Sprintf("%s/temp/config.json",project_dir)
	_, err = os.Stat(dst_dir)
	if err != nil {
		if os.IsNotExist(err) {
			os.Remove(dst_dir)
		}
	} else {
		os.Remove(dst_dir)
	}
	file, err := os.OpenFile(dst_dir, os.O_CREATE|os.O_RDWR, 0777)
	defer func ()  {
		err := file.Close()
		sing2catconfig.Logger_caller("File has been closed!",err)
	}()
	if err != nil{
		sing2catconfig.Logger_caller("Create file failed",err)
		return err
	}
	_,err = file.WriteString(string(config_file))
	if err != nil{
		sing2catconfig.Logger_caller("Write config failed!",err)
		return err
	}
	return nil
}

func Get_map_value(content map[string]interface{}, keys ...string)(any,error){
	var value any
	// 逐级获得字典的值
	for i,key := range(keys){
		if (i != len(keys) - 1){
			content = content[key].(map[string]interface{})
		}else{
			value = content[key]
		}
		
	}
	if value == nil{
		err := errors.New("fetch map value failed")
		return nil,err
	}
	return value,nil
}