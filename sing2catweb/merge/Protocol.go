package merge

import (
	"fmt"
	"regexp"
	initial "sing2catweb/Initial"
	"sync"

	"github.com/bitly/go-simplejson"
	"github.com/huandu/go-clone"
)

func formatTag(node_name string,lock *sync.RWMutex) string {
	lock.Lock()
	defer lock.Unlock()
	// 定义返回的标签名type
	var tag_name string
	// 获取国家列表
	countries := initial.GetValue("template").(*simplejson.Json).Get("country").MustArray()	
	for index,country := range(countries){
		// 创建正则匹配式
		reg := regexp.MustCompile(country.(string))	
		if reg.MatchString(node_name){
			// 匹配到则去全局变量查看这个国家出现的次数
			country_count := initial.GetValue(country.(string))
			if country_count == nil{
				// 为空则设置1
				tag_name = country.(string)+"-"+fmt.Sprint(1)
				initial.SetValue(country.(string),1)
			}else{
				// 不为空则在基础上加1
				// 首先加1的原因是因为一开始没有这个值,设置为1,之后再遇到这个值,因为是首先生成tag名
				// 此时计数是1,会和之前的重合,所以先加1
				tag_name = country.(string)+"-"+fmt.Sprint(country_count.(int)+1)
				initial.SetValue(country.(string),country_count.(int)+1)
			}
			return tag_name
		}else{
			// 与上面的逻辑一样,只是变成未知区域,一般是遍历整个表也没有之后才会进入未知区域
			// 所以才用索引判断
			if index == len(countries) - 1{
				unknown_count := initial.GetValue(country.(string))
				if unknown_count == nil{
					tag_name = country.(string) + "-" + fmt.Sprint(1)
					initial.SetValue(country.(string),1)
				}else{
					tag_name = country.(string) + "-" + fmt.Sprint(unknown_count.(int)+1)
					initial.SetValue(country.(string),1 + unknown_count.(int))
				}
			}
		}
	}
	return tag_name
}
func FormatNode(node map[string]interface{},lock *sync.RWMutex) *simplejson.Json{
	protocol_type := node["type"]
	var new_node *simplejson.Json
	switch protocol_type {
	case "vmess":
		new_node = clone.Clone(initial.GetValue("template").(*simplejson.Json).Get("outbounds").Get("vmess")).(*simplejson.Json)
		tag_name := formatTag(node["name"].(string),lock)
		new_node.Set("tag",tag_name)
		new_node.Set("server",node["server"])
		new_node.Set("server_port",int(node["port"].(int)))
		new_node.Set("uuid",node["uuid"])
		new_node.SetPath([]string{"transport","type"},node["network"])
		new_node.SetPath([]string{"transport","path"},node["ws-path"])
		new_node.SetPath([]string{"transport","headers"},node["ws-headers"])
	case "ss":
		new_node = clone.Clone(initial.GetValue("template").(*simplejson.Json).Get("outbounds").Get("ss")).(*simplejson.Json)
		tag_name := formatTag(node["name"].(string),lock)
		new_node.Set("tag",tag_name)
		new_node.Set("server",node["server"])
		new_node.Set("server_port",int(node["port"].(int)))
		new_node.Set("method",node["cipher"])
		new_node.Set("password",node["password"])
	case "trojan":
		new_node = clone.Clone(initial.GetValue("template").(*simplejson.Json).Get("outbounds").Get("trojan")).(*simplejson.Json)
		tag_name := formatTag(node["name"].(string),lock)
		new_node.Set("tag",tag_name)
		new_node.Set("server",node["server"])
		new_node.Set("server_port",int(node["port"].(int)))
		new_node.SetPath([]string{"tls","server_name"},node["sni"])
		new_node.Set("password",node["password"])
	}
	return new_node
}