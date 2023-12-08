package main

import (
	"json-to-go"
	"strconv"
	"strings"
	"syscall/js"
)

// 需要在ide里设置os和arch
// go to  Settings (Preferences) | Go | Vendoring & Build Tags  and then select  OS  ->  js  and  ARCH  ->  wasm.
func main() {
	js.Global().Set("JsonToGoGen", js.FuncOf(JsonToGoGen))
	signal := make(chan struct{})
	<-signal
}

func JsonToGoGen(this js.Value, args []js.Value) interface{} {
	config := core.Config{}
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
	commentStr := getStringVue(jsonValue, "comment")
	comment, _ := strconv.Atoi(commentStr)
	config.Comment = comment
	pointerFlag := getStringVue(jsonValue, "pointerFlag")
	if pointerFlag == "true" {
		config.PointerFlag = true
	}
	nestFlag := getStringVue(jsonValue, "nestFlag")
	if nestFlag == "true" {
		config.NestFlag = true
	}
	generate, err := core.Generate(jsonStr, &config)
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
