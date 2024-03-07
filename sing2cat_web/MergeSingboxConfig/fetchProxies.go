package mergesingboxconfig

import (
	"errors"
	"fmt"
	"regexp"
	sing2catconfig "sing2cat_web/Sing2catConfig"
	"sync"

	"github.com/gocolly/colly"
	"gopkg.in/yaml.v3"
)

func format_url() []interface{} {
	// 读取配置文件获得url列表
	content,err := sing2catconfig.Get_value("config")
	if err != nil {
		sing2catconfig.Logger_caller("Get config msg failed!",err)
		panic("Get config msg failed!")
	}
	urls := content.(map[string]interface{})["url"].([]interface{})
	if len(urls) == 0 {
		err = errors.New("get url list failed")
		sing2catconfig.Logger_caller("Unknown Error",err)
		panic("Get url list failed!")
	}
	// 对url进行检查,如果没用clash标签则补上
	for index, url := range urls {
		// 创建正则匹配器,根据?或者&切割
		reg := regexp.MustCompile(`\?|&`)
		parameters := reg.Split(url.(string), -1)
		for _, paparameter := range parameters {
			// 对切割后的参数进行检查,看是否有clash
			para_reg := regexp.MustCompile("clash")
			// 并将检查结果赋给clash_tag
			clash_tag := para_reg.MatchString(paparameter)
			// 如果有clash则退出参数循环
			if clash_tag {
				break
			} else {
				// 没用clash则补上
				if index == len(parameters)-1 {
					urls[index] = url.(string)+ "&flag=clash"
				}
			}
		}
	}
	// 返回检查后的url
	return urls
}
func fetchProxies() ([]map[string]interface{},error){
	// 创建异步锁,避免设置地区tag时序号混乱
	var lock sync.RWMutex
	var country_map = make(map[string]int)
	// 最终返回的节点切片
	proxies := []map[string]interface{}{}
	// 获取整理后的urls
	urls := format_url()
	// nodes_num会接受每个异步函数获取到的节点数,缓存默认为url的个数,避免阻塞
	proxies_num := make(chan int,len(urls))
	// channel用于接收每个节点的具体信息
	proxy_channel := make(chan map[string]interface{},100)
	// 判断通道是否开启并关闭通道,此步是保险步骤
	defer func() {
		_,ok := <- proxy_channel
		if ok {
			close(proxy_channel)
		}
	}()
	defer func() {
		_,ok := <- proxies_num
		if ok {
			close(proxies_num)
		}
	}()

	// 异步爬取数据,创建colly对象
	c := colly.NewCollector(colly.Async(true),colly.MaxDepth(len(urls)))
	
	// 设置请求头
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	
	// 得到响应的回调函数
	c.OnResponse(func(r *colly.Response) {
		// content是为了yaml解析对象创建的变量
		content := map[string]interface{}{}
		// yaml解析后的结果放入content中
		if err := yaml.Unmarshal(r.Body,&content);err!=nil{
			msg := fmt.Sprintf("Parse %s proxies yaml failed!",r.Request.URL)
			sing2catconfig.Logger_caller(msg,err)
			proxies_num <- 0
			return
		}
		// 如果解析到的内容没有proxies则退出函数
		if content["proxies"] == nil{
			err := errors.New("get proxies msg failed")
			msg := fmt.Sprintf("Get %s proxies msg failed!",r.Request.URL)
			sing2catconfig.Logger_caller(msg,err)
			proxies_num <- 0
			return
		}
		// temp_nodes是yaml状态的节点信息,之后会遍历送入处理函数返回json形式的
		temp_proxies := content["proxies"].([]interface{})
		// 向nodes_num发送节点数量信息
		proxies_num <- len(temp_proxies)
		// 处理yaml节点发送给channel通道
		for _,temp_proxy := range(temp_proxies){
			proxy_json,err := Format_proxy(temp_proxy.(map[string]interface{}),&lock,country_map)
			if err != nil{
				sing2catconfig.Logger_caller("Marshal proxy to json failed!",err)
				break
			}
			proxy_channel <- proxy_json
		}
	})
	
	// 得到错误的回调函数
	c.OnError(func(r *colly.Response, err error) {
		msg := fmt.Sprintf("Fetch %s proxies yaml failed!",r.Request.URL)
		sing2catconfig.Logger_caller(msg,err)
		// 失败则没有节点信息
		proxies_num <- 0
	})
	
	// 遍历url爬取yaml文件
	for _,url := range(urls){
		c.Visit(url.(string))
	}

	// finished_count用于记录url的结果,不论失败还是成功
	finished_count := 0
	// proxies_num_count用于记录从proxies_num通道记录的不同url的节点数量之和
	proxies_num_count := 0
	// proxy_append_count用于记录已经添加了多少个节点的信息
	proxy_append_count := 0
	// 自定义错误
	var err error
	
	// 遍历channel通道
	for {
		// 首先获得总的节点数量,所以先看url的访问情况,并从中获取每个url的节点数量并求和
		if finished_count < len(urls){
			// 对节点数量求和并记录一次url访问记录,在nodes_num有记录之前是会堵塞的
			proxies_num_count = proxies_num_count + <- proxies_num
			finished_count += 1
			// 一旦url访问完成说明nodes_num的记录结果已经求和完成了
			if finished_count == len(urls){
				close(proxies_num)
				// 如果节点数量是0,则没有意义往下进行,直接panic了
				if proxies_num_count == 0{
					err = errors.New("get proxies msg failed")
					sing2catconfig.Logger_caller("url or network Error",err)
					close(proxy_channel)
					return proxies,err
				}
			}
		}
		// 如果添加次数小于节点数量,则添加并记录
		if proxy_append_count < proxies_num_count{
			proxies = append(proxies, <- proxy_channel)
			proxy_append_count += 1
			// 添加完成之后关闭通道打断循环
			if proxy_append_count == proxies_num_count{
				close(proxy_channel)
				break
			}
		}
	}
	// 等待colly生命周期完成
	c.Wait()
	
	return proxies,nil
}

func select_outbound(tags []string) (map[string]interface{},error){
	// 获取模板信息
	content,err := sing2catconfig.Get_value("template")
	if err != nil{
		return nil,err
	}
	// 转换为simpleJson对象
	select_map,err := Get_map_value(content.(map[string]interface{}),"outbounds","select")
	if err != nil {
		sing2catconfig.Logger_caller("get select failed!",err)
		return nil,err
	}
	// 将模板中outbounds的自定义出站tag添加进去
	template_tags := select_map.(map[string]interface{})["outbounds"].([]interface{}) 
	for _,template_tag := range(template_tags){
		tags = append(tags, template_tag.(string))
	}
	
	// 设置选择节点的出站标签
	select_map.(map[string]interface{})["outbounds"] = tags
	return select_map.(map[string]interface{}),nil
}

func auto_outbound(tags []string) (map[string]interface{},error){
	// 获取模板信息
	content,err := sing2catconfig.Get_value("template")
	if err != nil{
		return nil,err
	}
	// 转换为simpleJson对象
	auto_map,err := Get_map_value(content.(map[string]interface{}),"outbounds","auto")
	if err != nil {
		sing2catconfig.Logger_caller("get auto outbounds failed!",err)
		return nil,err
	}
	// 将模板auto中outbounds的自定义出站tag添加进去
	template_tags := auto_map.(map[string]interface{})["outbounds"].([]interface{}) 
	for _,template_tag := range(template_tags){
		tags = append(tags, template_tag.(string))
	}
	// 设置自动节点的出站
	auto_map.(map[string]interface{})["outbounds"] = tags
	return auto_map.(map[string]interface{}),nil
}
func Merge_outbounds() ([]map[string]interface{},error){
	proxies,err := fetchProxies()
	if err != nil{
		return proxies,err
	}
	// 从节点中获得标签用于生成auto和select出站
	tags := make([]string,len(proxies))
	for index,proxy := range(proxies){
		tags[index] = proxy["tag"].(string)
	}
	// 给节点列表添加选择出站
	select_proxy,err := select_outbound(tags)
	if err != nil{
		sing2catconfig.Logger_caller("Get select outbounds failed!",err)
		return proxies,err
	}
	proxies = append(proxies, select_proxy)
	
	// 获取配置文件
	config,err := sing2catconfig.Get_value("config")
	if err != nil{
		sing2catconfig.Logger_caller("Get Config msg failed!",err)
		return proxies,err
	}
	// 获取rule_set
	rule_sets := config.(map[string]interface{})["rule_set"].([]interface{})
	for _,rule := range(rule_sets){
		// 如果china标签为否,说明是连国外的网站,为其生成selector类型出站
		if !rule.(map[string]interface{})["value"].(map[string]interface{})["china"].(bool){
			// 生成selector类型出站
			rule_set_select_outbound,err := select_outbound(tags)
			if err != nil{
				sing2catconfig.Logger_caller("Get select outbounds failed!",err)
				return proxies,err
			}
			// 生成selector类型出站成功则添加进出站列表
			rule_set_select_outbound["tag"] = rule.(map[string]interface{})["label"].(string)+"-select"
			proxies = append(proxies, rule_set_select_outbound)
		}
	}
	// 添加自动出站节点
	auto_outbound,err := auto_outbound(tags)
	if err != nil {
		sing2catconfig.Logger_caller("Get auto outbound failed",err)
		return proxies,err	
	}
	proxies = append(proxies, auto_outbound)
	// 获取模板信息
	content,err := sing2catconfig.Get_value("template")
	if err != nil{
		sing2catconfig.Logger_caller("Get Template failed!",err)
		return proxies,err	
	}
	// 获取模板信息的自定义出站信息
	custom_outbounds := content.(map[string]interface{})["outbounds"].(map[string]interface{})["custom_outbound"]
	// 遍历自定义出站信息并添加
	for _,custom_outbound := range(custom_outbounds.([]interface{})){
		proxies = append(proxies, custom_outbound.(map[string]interface{}))
	}		
	return proxies,nil
}