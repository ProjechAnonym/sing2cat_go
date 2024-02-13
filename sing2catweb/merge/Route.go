package merge

import (
	initial "sing2catweb/Initial"

	"github.com/bitly/go-simplejson"
	"github.com/huandu/go-clone"
)

func formatRuleSetSource() (*simplejson.Json,string){
	// 默认模板有geoip和geosite的下载路径,所以需要保留
	default_rule_set := clone.Clone(initial.GetValue("template").(*simplejson.Json).GetPath("route", "rule_set").MustArray()).([]interface{})
	// 获取自己的规则集
	custom_rule_set := clone.Clone(initial.GetValue("config").(*simplejson.Json).Get("rule_set").MustArray()).([]interface{})
	// 由于自己的规则集格式与singbox的配置格式不同,所以创建一个字典用于生成匹配singbox的配置
	rule_set_source := make(map[string]string)
	// 转换为json
	rule_set_source_json := simplejson.New()
	// 遍历自定义的规则集
	for _,rule := range custom_rule_set {
		// 这个for循环其实只是为了取自定义规则集的键
		for key := range(rule.(map[string]interface{})){
			// 根据自定义规则集的type确定不同类型
			switch rule.(map[string]interface{})[key].(map[string]interface{})["type"]{
			case "local":
				rule_set_source = map[string]string{"type":"local","tag":key,"format":"binary","path":rule.(map[string]interface{})[key].(map[string]interface{})["path"].(string)}
			case "remote":
				rule_set_source = map[string]string{"type":"remote","tag":key,"format":"binary","url":rule.(map[string]interface{})[key].(map[string]interface{})["url"].(string),"download_detour": "select"}
			}
			// 将设置好的字典赋给json对象
			rule_set_source_json.Set(key,rule_set_source)
			// 这步感觉有点多余,但是懒得改了
			default_rule_set = append(default_rule_set, rule_set_source_json.Get(key))
		}	
	}
	// 最后返回一个存储rule_set信息的json对象,将键值信息一并返回,避免未来改键后续代码报错
	default_rule_set_json := simplejson.New()
	default_rule_set_json.Set("rule_set",default_rule_set)
	return default_rule_set_json,"rule_set"
}

func formatRuleSet() (*simplejson.Json ,string){
	// 获取原本的route的规则
	original_rules := clone.Clone(initial.GetValue("template").(*simplejson.Json).GetPath("route","rules").MustArray()).([]interface{})
	// 获取自定义的规则
	custom_rules := clone.Clone(initial.GetValue("config").(*simplejson.Json).Get("rule_set").MustArray()).([]interface{})
	// 基础的dns,屏蔽规则
	base_rules := original_rules[0:3]
	// 获取分流规则
	shunt_rules := clone.Clone(original_rules[len(original_rules)-2:]).([]interface{})
	// 此处逻辑于上面相同
	for _,rule := range(custom_rules){
		for key := range(rule.(map[string]interface{})){
			switch rule.(map[string]interface{})[key].(map[string]interface{})["china"]{
			case true:
				base_rules = append(base_rules, map[string]string{"rule_set":key,"outbound":"direct"})
			case false:
				base_rules = append(base_rules, map[string]string{"rule_set":key,"outbound":key+"-select"})
			}
		}
	}
	base_rules = append(base_rules, shunt_rules...)
	base_rules_json := simplejson.New()
	base_rules_json.Set("rules",base_rules)
	return base_rules_json,"rules"
}

func MergeRoute() *simplejson.Json{
	original_route := clone.Clone(initial.GetValue("template").(*simplejson.Json).Get("route")).(*simplejson.Json)
	rule_set,rule_set_key := formatRuleSetSource()
	rules,rules_key := formatRuleSet()
	// route有规则集和规则,需要分开设置
	original_route.Set(rule_set_key,rule_set.Get(rule_set_key))
	original_route.Set(rules_key,rules.Get(rules_key))
	return original_route
}