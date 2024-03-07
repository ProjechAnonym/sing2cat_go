# sing2cat_go
## 基本功能
[八音猫功能](https://github.com/ProjechAnonym/sing2cat/blob/main/README.md)
## 项目路径配置
```
E:\MYPROJECT\DEMO
│  sing2cat_web.exe
│
├─config
│  └─sing2cat
│          config.yaml
│          template.yaml
│    
├─log
│      sing2cat-info.log
│  
└─temp
        config.json
```
DEMO为项目文件夹,config用于存储八音猫的配置文件和模板文件,log用于存放日志,temp目录则存放生成的sing-box配置文件。
## 配置文件config.json
主要是机场链接和规则集链接,支持多个机场链接聚合,根据国家进行分类,国家列表在template.json文件中,请确保未知地区在最后一个!!
### 规则集相关
规则集配置的标签参考sing-box官方并作出简化,只需要指明规则集是本地还是远程,本地需要指明规则集文件的本地绝对路径,远程规则集需要指明url链接。

china标签表示的是这个规则集是国外还是国内,比如谷歌服务相关就把china标签填false,需要直连的比如b站就将china标签填true。规则集列表可以为空。

需要代理的规则集会自动生成一个专属的select节点,比如gpt规则集会生成一个不同于默认的select节点的1select节点,这样可以保证gpt的出站走指定代理避免连不上的问题。

远程规则集下载时默认走默认生成的select节点,另外为了性能,规则集均采用sing-box官方示例的二进制文件模式,我不希望有人去更改这样的设计,当然你是大佬魔改我的代码我也不能怎么样,反正我的注释也有
## 代码运行逻辑&如何添加新的协议
程序通过读取template.json文件生成配置信息,因此添加新的协议需要在template.json文件中的outbounds添加你新的协议的模板,之后在Merge文件夹下的Protocol.go中FormatNode函数中swicth选择分支中加入新的协议。

template.json中有些模板代码读取之后是不会改变的,你可以随意修改,dns,log,inbounds,experimental都是原封不动抄到新的config.json的。这样做的好处是你可以根据客户端的需要定制你的配置文件,比如一些不支持tun模式的设备你就可以自定义template.json的入站配置适配这种设备。
## 关于编译
go语言非常好编译,你只需要下载我的代码,配置好go环境,然后进入项目文件夹中,命令行go build就行了
