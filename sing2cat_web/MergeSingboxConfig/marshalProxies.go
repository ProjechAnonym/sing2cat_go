package mergesingboxconfig

import (
	"errors"
	"fmt"
	"regexp"
	sing2catconfig "sing2cat_web/Sing2catConfig"
	"sync"
)

func rename_tag(proxy_name string,lock *sync.RWMutex,country_map map[string]int) string {
	lock.Lock()
	defer lock.Unlock()
	// 定义返回的标签名type
	var tag_name string
	// 获取模板信息
	content,err := sing2catconfig.Get_value("template")
	if err != nil{
		sing2catconfig.Logger_caller("Get Template failed!",err)
		panic("Get Template failed!")
	}
	// 获取国家列表
	countries := content.(map[string]interface{})["country"].([]interface{})
	for index,country := range(countries){
		// 创建正则匹配式
		reg := regexp.MustCompile(country.(string))
		if reg.MatchString(proxy_name){
			// 匹配到则去国家计数字典查看这个国家出现的次数
			country_count := country_map[country.(string)]
			if country_count == 0{
				// 为空则设置1并设置tag_name
				tag_name = country.(string)+"-"+fmt.Sprint(1)
				country_map[country.(string)] = 1
			}else{
				// 不为空则在原基础上加1
				// 文件名先加1的原因是因为一开始没有这个值,设置为1,之后再遇到这个值,因为是首先生成tag名
				// 此时计数是1,会和之前的重合,所以先加1
				tag_name = country.(string)+"-"+fmt.Sprint(country_count+1)
				country_map[country.(string)] = country_count+1
			}
			return tag_name
		}else{
			// 与上面的逻辑一样,只是变成未知区域,一般是遍历整个表也没有之后才会进入未知区域
			// 用索引数判断是否遍历到最后
			if index == len(countries) - 1{
				unknown_count := country_map[country.(string)]
				if unknown_count == 0{
					tag_name = country.(string) + "-" + fmt.Sprint(1)
					country_map[country.(string)] = 1
				}else{
					tag_name = country.(string) + "-" + fmt.Sprint(unknown_count+1)
					country_map[country.(string)] = 1 + unknown_count
				}
			}
		}
	}
	return tag_name
}
func Format_proxy(proxy_map map[string]interface{},lock *sync.RWMutex,country_count map[string]int) (map[string]interface{},error){
	// 获取协议类型
	protocol_type := proxy_map["type"]

	switch protocol_type {
		case "vmess":
			// 获取模板信息
			content,err := sing2catconfig.Get_value("template")
			if err != nil{
				sing2catconfig.Logger_caller("Get Template failed!",err)
				panic("Get Template failed!")
			}
			// 转为map
			proxy_vmess,err := Get_map_value(content.(map[string]interface{}),"outbounds","vmess")
			if err != nil {
				return nil,err
			}
			tag_name := rename_tag(proxy_map["name"].(string),lock,country_count)
			proxy_vmess.(map[string]interface{})["tag"] = tag_name
			proxy_vmess.(map[string]interface{})["server"] = proxy_map["server"]
			proxy_vmess.(map[string]interface{})["server_port"] = int(proxy_map["port"].(int))
			proxy_vmess.(map[string]interface{})["uuid"] = proxy_map["uuid"]
			proxy_vmess.(map[string]interface{})["transport"].(map[string]interface{})["type"] = proxy_map["network"]
			proxy_vmess.(map[string]interface{})["transport"].(map[string]interface{})["path"] = proxy_map["ws-path"]
			proxy_vmess.(map[string]interface{})["transport"].(map[string]interface{})["headers"] = proxy_map["ws-headers"]
			return proxy_vmess.(map[string]interface{}),nil
		case "ss":
			// 获取模板信息
			content,err := sing2catconfig.Get_value("template")
			if err != nil{
				sing2catconfig.Logger_caller("Get Template failed!",err)
				panic("Get Template failed!")
			}
			// 转为map
			proxy_ss,err := Get_map_value(content.(map[string]interface{}),"outbounds","ss")
			if err != nil {
				return nil,err
			}
			tag_name := rename_tag(proxy_map["name"].(string),lock,country_count)
			proxy_ss.(map[string]interface{})["tag"] = tag_name
			proxy_ss.(map[string]interface{})["server"] = proxy_map["server"]
			proxy_ss.(map[string]interface{})["server_port"] = int(proxy_map["port"].(int))
			proxy_ss.(map[string]interface{})["method"] = proxy_map["cipher"]
			proxy_ss.(map[string]interface{})["password"] = proxy_map["password"]
			return proxy_ss.(map[string]interface{}),nil
		case "trojan":
			// 获取模板信息
			content,err := sing2catconfig.Get_value("template")
			if err != nil{
				sing2catconfig.Logger_caller("Get Template failed!",err)
				panic("Get Template failed!")
			}
			// 转为map
			proxy_trojan,err := Get_map_value(content.(map[string]interface{}),"outbounds","trojan")
			if err != nil {
				return nil,err
			}
			tag_name := rename_tag(proxy_map["name"].(string),lock,country_count)
			proxy_trojan.(map[string]interface{})["tag"] = tag_name
			proxy_trojan.(map[string]interface{})["server"] = proxy_map["server"]
			proxy_trojan.(map[string]interface{})["server_port"] = int(proxy_map["port"].(int))
			proxy_trojan.(map[string]interface{})["tls"].(map[string]interface{})["server_name"] = proxy_map["sni"]
			proxy_trojan.(map[string]interface{})["password"] = proxy_map["password"]
			return proxy_trojan.(map[string]interface{}),nil
		}
	msg := fmt.Sprintf("protocol %s is not in the template",protocol_type)
	err := errors.New(msg)
	return nil,err
}