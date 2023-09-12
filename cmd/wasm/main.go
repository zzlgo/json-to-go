package main

import (
	"json2go"
	"strings"
	"syscall/js"
)

// 需要在ide里设置os和arch
// go to  Settings (Preferences) | Go | Vendoring & Build Tags  and then select  OS  ->  js  and  ARCH  ->  wasm.
func main() {
	js.Global().Set("Json2GoGen", js.FuncOf(Json2GoGen))
	signal := make(chan struct{})
	<-signal
}

func Json2GoGen(this js.Value, args []js.Value) interface{} {
	config := json2go.Config{}
	jsonValue := args[0]
	jsonStr := getStringVue(jsonValue, "jsonStr")
	var tags []string
	for _, t := range []string{"jsonTag", "bsonTag", "mapstructureTag", "customTag"} {
		tagValue := getStringVue(jsonValue, t)
		if tagValue != "" {
			if strings.Contains(tagValue, ",") {
				split := strings.Split(tagValue, ",")
				tags = append(tags, split...)
			} else {
				tags = append(tags, tagValue)
			}
		}
	}
	config.Tags = tags
	commentFlag := getStringVue(jsonValue, "commentFlag")
	if commentFlag == "true" {
		config.CommentFlag = true
	}
	pointerFlag := getStringVue(jsonValue, "pointerFlag")
	if pointerFlag == "true" {
		config.PointerFlag = true
	}
	generate, err := json2go.Generate(jsonStr, &config)
	if err != nil {
		return map[string]interface{}{
			"code":    500,
			"message": generate,
		}
	}
	return map[string]interface{}{
		"code": 0,
		"data": generate,
	}
}

func getStringVue(jsonValue js.Value, key string) string {
	value := jsonValue.Get(key)
	if !value.IsUndefined() && !value.IsNull() && !value.IsNaN() {
		return value.String()
	}
	return ""
}
