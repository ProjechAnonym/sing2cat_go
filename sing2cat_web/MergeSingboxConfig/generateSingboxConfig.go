package mergesingboxconfig

import (
	"errors"
	sing2catconfig "sing2cat_web/Sing2catConfig"
	"sync"

	"github.com/bitly/go-simplejson"
)

func GenerateConfigJson() (error){
	template,err := sing2catconfig.Get_value("template")
	if err != nil{
		sing2catconfig.Logger_caller("Get template failed!",err)
		return err
	}
	outbound_channel := make(chan map[string]interface{}, 50)
	outbounds := []map[string]interface{}{}
	var jobs sync.WaitGroup
	jobs.Add(1)
	// 获取固定信息
	log,err := Get_map_value(template.(map[string]interface{}),"log")
	if err != nil{
		sing2catconfig.Logger_caller("Get log msg failed!",err)
		return err
	}
	dns,err := Get_map_value(template.(map[string]interface{}),"dns")
	if err != nil{
		sing2catconfig.Logger_caller("Get dns msg failed!",err)
		return err
	}
	inbounds,err := Get_map_value(template.(map[string]interface{}),"inbounds")
	if err != nil{
		sing2catconfig.Logger_caller("Get inbounds msg failed!",err)
		return err
	}
	experimental,err := Get_map_value(template.(map[string]interface{}),"experimental")
	if err != nil{
		sing2catconfig.Logger_caller("Get experimental msg failed!",err)
		return err
	}
	// 获取会变化的信息,出站和路由
	route,err := Merge_route()
	if err != nil{
		sing2catconfig.Logger_caller("Get route failed!",err)
		return err
	}
	sing2catconfig.Logger_caller("Generate route completed!",nil)
	go func() {
		defer jobs.Done()
		defer close(outbound_channel)
		outbounds, err := Merge_outbounds()
		if err != nil {
			return
		}
		sing2catconfig.Logger_caller("Fetch proxies completed!",nil)
		for _, outbound := range outbounds {
			outbound_channel <- outbound
		}
	}()
	// 设置json
	sing_box_config := simplejson.New()
	sing_box_config.Set("log", log)
	sing_box_config.Set("dns", dns)
	sing_box_config.Set("inbounds", inbounds)
	sing_box_config.Set("route", route)
	sing_box_config.Set("experimental", experimental)

	for outbound := range outbound_channel {
		outbounds = append(outbounds, outbound)
	}
	jobs.Wait()
	if len(outbounds) == 0{
		err = errors.New("fetch proxies failed")
		return err
	}
	sing_box_config.Set("outbounds", outbounds)
	config_file, _ := sing_box_config.EncodePretty()
	err = Write_config_file(config_file)
	if err != nil{
		return err
	}
	return nil
	
}