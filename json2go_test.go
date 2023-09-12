package json2go

import (
	"strings"
	"testing"
)

// web使用的测试case
func TestGenerate(t *testing.T) {
	jsonStr := `{
	// 1.支持单行备注提取，可在上一行或行尾
	"k1": "v1",
	"k2": "v2", // 行尾备注
	"k3": {
		"k3_1": {
			"k3_1_1": "" // 递归解析备注
		}
	},
	"k4": [{ // 数组类型备注
		"k41": ""
	}, {
		"k41": "" // 备注合并
	}],
	// 2.支持属性名格式化，支持中文
	"地址": "山东枣庄",
	"地质": "中文解析后重名了，自动数字编号",
	"中文++))": "自动忽略不识别的字符",
	"id": "golint代码检查优化",
	"user_name": "",
	"3只松鼠": "",
	// 3.支持属性类型推断，单独属性，数组内属性，同名对象属性
	"float": 1.15,
	"arrayFloat": [1, 2.1],
	"arrayObjFloat1": [{
		"cost": 10
	}, {
		"cost": 10.2
	}],
	"arrayObjFloat2": [
		[{
			"cost": 10
		}, {
			"cost": 10.2
		}]
	],
	// 4.支持对象属性合并，重名的所有对象，包括数组内所有对象
	"questionArray": [{
			"type": "select",
			"props": {
				"info": ""
			},
			"options": [],
			"other": null
		},
		{
			"type": "select",
			"props": {
				"info": "",
				"value": ""
			},
			"options": [{
				"label": "",
				"value": ""
			}],
			"other": "",
			"required": false
		}
	]
}`
	tags := []string{"json"}
	generate, err := Generate(jsonStr, &Config{
		Tags:        tags,
		CommentFlag: true,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	res := `type AutoGenerated struct {
	// 1.支持单行备注提取，可在上一行或行尾
	K1 string |json:"k1"|
	// 行尾备注
	K2 string |json:"k2"|
	K3 K3     |json:"k3"|
	// 数组类型备注
	K4 []K4 |json:"k4"|
	// 2.支持属性名格式化，支持中文
	Dz       string |json:"地址"|
	Dz1      string |json:"地质"|
	Zw       string |json:"中文++))"|
	ID       string |json:"id"|
	UserName string |json:"user_name"|
	Threezss string |json:"3只松鼠"|
	// 3.支持属性类型推断，单独属性，数组内属性，同名对象属性
	Float          float64            |json:"float"|
	ArrayFloat     []float64          |json:"arrayFloat"|
	ArrayObjFloat1 []ArrayObjFloat1   |json:"arrayObjFloat1"|
	ArrayObjFloat2 [][]ArrayObjFloat2 |json:"arrayObjFloat2"|
	// 4.支持对象属性合并，重名的所有对象，包括数组内所有对象
	QuestionArray []QuestionArray |json:"questionArray"|
}

type K3 struct {
	K31 K31 |json:"k3_1"|
}

type K31 struct {
	// 递归解析备注
	K311 string |json:"k3_1_1"|
}

type K4 struct {
	// 备注合并
	K41 string |json:"k41"|
}

type ArrayObjFloat1 struct {
	Cost float64 |json:"cost"|
}

type ArrayObjFloat2 struct {
	Cost float64 |json:"cost"|
}

type QuestionArray struct {
	Type     string    |json:"type"|
	Props    Props     |json:"props"|
	Options  []Options |json:"options"|
	Other    string    |json:"other"|
	Required bool      |json:"required"|
}

type Props struct {
	Info  string |json:"info"|
	Value string |json:"value"|
}

type Options struct {
	Label string |json:"label"|
	Value string |json:"value"|
}`
	generate = strings.ReplaceAll(generate, "`", "|")
	if generate != res {
		t.Fatalf(generate)
	}
}

// 测试备注
func TestGenerateComment(t *testing.T) {
	jsonStr := `{
	// 1.支持单行备注提取，可在上一行或行尾
	"k1": "v1",
	"k2": "v2", // 行尾备注
	"k3": {
		"k3_1": {
			"k3_1_1": "" // 递归解析备注
		}
	},
	"k4": [{ // 数组类型备注
		"k41": ""
	}, {
		"k41": "" // 备注合并
	}]
}`
	tags := []string{"json"}
	generate, err := Generate(jsonStr, &Config{
		Tags:        tags,
		CommentFlag: true,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	res := `type AutoGenerated struct {
	// 1.支持单行备注提取，可在上一行或行尾
	K1 string |json:"k1"|
	// 行尾备注
	K2 string |json:"k2"|
	K3 K3     |json:"k3"|
	// 数组类型备注
	K4 []K4 |json:"k4"|
}

type K3 struct {
	K31 K31 |json:"k3_1"|
}

type K31 struct {
	// 递归解析备注
	K311 string |json:"k3_1_1"|
}

type K4 struct {
	// 备注合并
	K41 string |json:"k41"|
}`
	generate = strings.ReplaceAll(generate, "`", "|")
	if generate != res {
		t.Fatalf(generate)
	}
}

// 测试属性的命名，支持中文，处理特殊字符
func TestGenerateName(t *testing.T) {
	jsonStr := `{
	"地址": "山东枣庄",
	"地质": "中文解析后重名了，自动数字编号",
	"中文++))": "自动忽略不识别的字符",
	"id": "golint代码检查优化",
	"user_name": "",
	"3只松鼠": ""
}`
	tags := []string{"json"}
	generate, err := Generate(jsonStr, &Config{
		Tags:        tags,
		CommentFlag: false,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	res := `type AutoGenerated struct {
	Dz       string |json:"地址"|
	Dz1      string |json:"地质"|
	Zw       string |json:"中文++))"|
	ID       string |json:"id"|
	UserName string |json:"user_name"|
	Threezss string |json:"3只松鼠"|
}`
	generate = strings.ReplaceAll(generate, "`", "|")
	if generate != res {
		t.Fatalf(generate)
	}
}

// 测试属性类型推断
func TestGenerateType(t *testing.T) {
	jsonStr := `{
	"float": 1.15,
	"arrayFloat": [1, 2.1],
	"arrayObjFloat1": [{
		"cost": 10
	}, {
		"cost": 10.2
	}],
	"arrayObjFloat2": [
		[{
			"cost": 10
		}, {
			"cost": 10.2
		}]
	]
}`
	tags := []string{"json"}
	generate, err := Generate(jsonStr, &Config{
		Tags:        tags,
		CommentFlag: true,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	res := `type AutoGenerated struct {
	Float          float64            |json:"float"|
	ArrayFloat     []float64          |json:"arrayFloat"|
	ArrayObjFloat1 []ArrayObjFloat1   |json:"arrayObjFloat1"|
	ArrayObjFloat2 [][]ArrayObjFloat2 |json:"arrayObjFloat2"|
}

type ArrayObjFloat1 struct {
	Cost float64 |json:"cost"|
}

type ArrayObjFloat2 struct {
	Cost float64 |json:"cost"|
}`
	generate = strings.ReplaceAll(generate, "`", "|")
	if generate != res {
		t.Fatalf(generate)
	}
}

// 测试对象的合并，同一个属性出现多个类型时，依赖类型判断来推断
func TestGenerateMergeObj(t *testing.T) {
	jsonStr := `{
	"questionArray": [{
			"type": "select",
			"props": {
				"info": ""
			},
			"options": [],
			"other": null
		},
		{
			"type": "select",
			"props": {
				"info": "",
				"value": ""
			},
			"options": [{
				"label": "",
				"value": ""
			}],
			"other": "",
			"required": false
		}
	]
}`
	tags := []string{"json"}
	generate, err := Generate(jsonStr, &Config{
		Tags:        tags,
		CommentFlag: true,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	res := `type AutoGenerated struct {
	QuestionArray []QuestionArray |json:"questionArray"|
}

type QuestionArray struct {
	Type     string    |json:"type"|
	Props    Props     |json:"props"|
	Options  []Options |json:"options"|
	Other    string    |json:"other"|
	Required bool      |json:"required"|
}

type Props struct {
	Info  string |json:"info"|
	Value string |json:"value"|
}

type Options struct {
	Label string |json:"label"|
	Value string |json:"value"|
}`
	generate = strings.ReplaceAll(generate, "`", "|")
	if generate != res {
		t.Fatalf(generate)
	}
}

// 测试空值，空数组的处理
func TestNullObj(t *testing.T) {
	jsonStr := `{
	"array": [{
		"options": [],
		"other": null
	}],
	"obj": {
		"options": [],
		"other": null
	},
	"options": [],
	"other": null
}`
	tags := []string{"json"}
	generate, err := Generate(jsonStr, &Config{
		Tags:        tags,
		CommentFlag: true,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	res := `type AutoGenerated struct {
	Array   []Array       |json:"array"|
	Obj     Obj           |json:"obj"|
	Options []interface{} |json:"options"|
	Other   interface{}   |json:"other"|
}

type Array struct {
	Options []interface{} |json:"options"|
	Other   interface{}   |json:"other"|
}

type Obj struct {
	Options []interface{} |json:"options"|
	Other   interface{}   |json:"other"|
}`
	generate = strings.ReplaceAll(generate, "`", "|")
	if generate != res {
		t.Fatalf(generate)
	}
}

// 测试输出结果中对象的顺序
func TestGenerateOrder(t *testing.T) {
	jsonStr := `{
	"中国": {
		"山东": {
			"济南": {},
			"青岛": {}
		},
		"浙江": {
			"杭州": {}
		}
	},
	"测试数组合并后的顺序": [{
		"山东": {
			"济南": {},
			"青岛": {}
		}
	}, {
		"山东": {
			"枣庄": {},
			"临沂": {}
		}
	}]
}`
	tags := []string{"json"}
	generate, err := Generate(jsonStr, &Config{
		Tags:        tags,
		CommentFlag: true,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	res := `type AutoGenerated struct {
	Zg         Zg           |json:"中国"|
	Csszhbhdsx []Csszhbhdsx |json:"测试数组合并后的顺序"|
}

type Zg struct {
	Sd Sd |json:"山东"|
	Zj Zj |json:"浙江"|
}

type Sd struct {
	Jn Jn |json:"济南"|
	Qd Qd |json:"青岛"|
	Zz Zz |json:"枣庄"|
	Ly Ly |json:"临沂"|
}

type Jn struct {
}

type Qd struct {
}

type Zz struct {
}

type Ly struct {
}

type Zj struct {
	Hz Hz |json:"杭州"|
}

type Hz struct {
}

type Csszhbhdsx struct {
	Sd Sd |json:"山东"|
}`
	generate = strings.ReplaceAll(generate, "`", "|")
	if generate != res {
		t.Fatalf(generate)
	}
}

// 测试入参是数组的情况
func TestGenerateArray(t *testing.T) {
	jsonStr := `[{
 	"id": 1,
 	"name": ""
 }, {
 	"id": 1,
 	"name": "",
 	"location": {
 		"province": "",
 		"city": ""
 	}
 }]`
	tags := []string{"json"}
	generate, err := Generate(jsonStr, &Config{
		Tags:        tags,
		CommentFlag: true,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	res := `type AutoGenerated struct {
	ID       int      |json:"id"|
	Name     string   |json:"name"|
	Location Location |json:"location"|
}

type Location struct {
	Province string |json:"province"|
	City     string |json:"city"|
}`
	generate = strings.ReplaceAll(generate, "`", "|")
	if generate != res {
		t.Fatalf(generate)
	}
}

// 测试使用指针
func TestPointer(t *testing.T) {
	jsonStr := `{
	"location": {
		"province": ""
	},
	"location1": [{
		"province": ""
	}],
	"location2": [
		[{
			"province": ""
		}]
	]
}`
	tags := []string{"json"}
	generate, err := Generate(jsonStr, &Config{
		Tags:        tags,
		CommentFlag: true,
		PointerFlag: true,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	res := `type AutoGenerated struct {
	Location  *Location      |json:"location"|
	Location1 []*Location1   |json:"location1"|
	Location2 [][]*Location2 |json:"location2"|
}

type Location struct {
	Province string |json:"province"|
}

type Location1 struct {
	Province string |json:"province"|
}

type Location2 struct {
	Province string |json:"province"|
}`
	generate = strings.ReplaceAll(generate, "`", "|")
	if generate != res {
		t.Fatalf(generate)
	}
}