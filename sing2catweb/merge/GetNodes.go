package merge

import (
	"errors"
	"regexp"
	initial "sing2catweb/Initial"
	"sync"

	"github.com/bitly/go-simplejson"
	"github.com/gocolly/colly"
	"github.com/huandu/go-clone"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func formatUrl() []interface{}{
	// 读取配置文件获得url列表
	config_content := initial.GetValue("config").(*simplejson.Json)
	urls,_ := config_content.Get("url").Array()
	// 对url进行检查,如果没用clash标签则补上
	for index,url := range(urls){
		// 创建正则匹配器,根据?或者&切割
		reg := regexp.MustCompile(`\?|&`)
		parameters := reg.Split(url.(string),-1)

		for _,paparameter := range(parameters){
			// 对切割后的参数进行检查,看是否有clash
			para_reg := regexp.MustCompile("clash")
			// 并将检查结果赋给clash_tag
			clash_tag := para_reg.MatchString(paparameter)
			// 如果有clash则退出参数循环			
			if clash_tag{
				break
			}else{
				// 没用clash则补上
				if index == len(parameters) - 1 {
					urls[index] = url.(string)+"&flag=clash"
				}			
			}			
		}
	}
	// 返回检查后的url
	return urls
}

func getNodes(Logger *zap.Logger) []*simplejson.Json{
	// 创建异步锁,避免设置地区tag时序号混乱
	var lock sync.RWMutex
	// 最终返回的节点切片
	nodes := []*simplejson.Json{}
	// 获取整理后的urls
	urls := formatUrl()
	if len(urls) == 0{
		return nodes
	}
	// nodes_num会接受每个异步函数获取到的节点数,缓存默认为url的个数,避免阻塞
	nodes_num := make(chan int,len(urls))
	// channel用于接收每个节点的具体信息
	channel := make(chan *simplejson.Json,100)
	// 判断通道是否开启并关闭通道,此步是保险步骤
	defer func() {
		_,ok := <- channel
		if ok {
			close(channel)
		}
	}()
	defer func() {
		_,ok := <- nodes_num
		if ok {
			close(nodes_num)
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
		err := yaml.Unmarshal(r.Body,&content)
		if err != nil{
			Logger.Error("节点yaml解析失败",zap.Error((err)))
		}
		// temp_nodes是yaml状态的节点信息,之后会遍历送入处理函数返回json形式的
		temp_nodes := content["proxies"].([]interface{})
		// 向nodes_num发送节点数量信息
		nodes_num <- len(temp_nodes)
		// 处理yaml节点发送给channel通道
		for _,temp_node := range(temp_nodes){
			channel <- FormatNode(temp_node.(map[string]interface{}),&lock)
		}
	})
	// 得到错误的回调函数
	c.OnError(func(r *colly.Response, err error) {
		if err != nil{
			Logger.Error("获取节点配置文件失败",zap.Error((err)))
		}
		// 失败则没有节点信息
		nodes_num <- 0
	})
	// 遍历url爬取yaml文件
	for _,url := range(urls){
		c.Visit(url.(string))
	}

	// finished_count用于记录url的结果,不论失败还是成功
	finished_count := 0
	// nodes_num_count用于记录从nodes_num通道记录的不同url的节点数量之和
	nodes_num_count := 0
	// node_append_count用于记录已经添加了多少个节点的信息
	node_append_count := 0
	// 遍历channel通道
	for {
		// 首先获得总的节点数量,所以先看url的访问情况,并从中获取每个url的节点数量并求和
		if finished_count < len(urls){
			// 对节点数量求和并记录一次url访问记录,在nodes_num有记录之前是会堵塞的
			nodes_num_count = nodes_num_count + <- nodes_num
			finished_count += 1
			// 一旦url访问完成说明nodes_num的记录结果已经求和完成了
			if finished_count == len(urls){
				close(nodes_num)
				// 如果节点数量是0,则没有意义往下进行,直接panic了
				if nodes_num_count == 0{
					err := errors.New("未能获得任何节点信息")
					Logger.Error("节点信息获取失败",zap.Error(err))
					close(channel)
					break
				}
			}
		}
		// 如果添加次数小于节点数量,则添加并记录
		if node_append_count < nodes_num_count{
			nodes = append(nodes, <- channel)
			node_append_count += 1
			// 添加完成之后关闭通道打断循环
			if node_append_count == nodes_num_count{
				close(channel)
				break
			}
		}
	}
	// 等待colly生命周期完成
	c.Wait()
	return nodes
}

func selectNode(tags []string) *simplejson.Json{
	// 生成的节点选择默认自动
	tags = append(tags, "auto")
	// 为防止引用造成的值修改,这里采用深拷贝的方法
	select_node := clone.Clone(initial.GetValue("template").(*simplejson.Json).GetPath("outbounds","select")).(*simplejson.Json)
	// 设置选择节点的出站标签
	select_node.Set("outbounds",tags)
	return select_node
}

func autoNode(tags []string) *simplejson.Json{
	// 为防止引用造成的值修改,这里采用深拷贝的方法
	auto_node := clone.Clone(initial.GetValue("template").(*simplejson.Json).GetPath("outbounds","auto")).(*simplejson.Json)
	// 设置自动节点的出站
	auto_node.Set("outbounds",tags)
	return auto_node
}

func MergeOutbounds(Logger *zap.Logger) []*simplejson.Json{
	// 首先获得节点信息
	nodes := getNodes(Logger)
	Logger.Info("节点获取",zap.String("状态","成功"))
	// 从节点中获得标签用于生成auto和select出站
	tags := make([]string,len(nodes))
	for index,node := range(nodes){
		tags[index] = node.Get("tag").MustString()
	}
	// 给节点列表添加选择出站
	nodes = append(nodes, selectNode(tags))
	// 查看ruleset
	rule_set := clone.Clone(initial.GetValue("config").(*simplejson.Json).Get("rule_set").MustArray()).([]interface{})
	for _,rule := range(rule_set){
		// rule_set是一个存储字典的列表,rule此时是一个字典,遍历字典的key
		for key := range(rule.(map[string]interface{})){
			// 如果ruleset的中国标签为否则表示是外网连接,为其生成select出站
			if !rule.(map[string]interface{})[key].(map[string]interface{})["china"].(bool){
				rule_set_select_node := selectNode(tags)
				rule_set_select_node.Set("tag",key+"-select")
				nodes = append(nodes, rule_set_select_node)
			}
		}
	}
	// 添加自动节点以及一些必要的节点
	nodes = append(nodes, autoNode(tags))
	nodes = append(nodes, initial.GetValue("template").(*simplejson.Json).GetPath("outbounds","direct"))
	nodes = append(nodes, initial.GetValue("template").(*simplejson.Json).GetPath("outbounds","dns_out"))
	nodes = append(nodes, initial.GetValue("template").(*simplejson.Json).GetPath("outbounds","block"))
	return nodes
}