package initial
var global_vars = make(map[string]interface{})

func GetValue(key string) interface{}{
	result := global_vars[key]
	return result
}
func SetValue(key string,value interface{}){
	global_vars[key] = value
}
