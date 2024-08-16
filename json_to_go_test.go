package core

import (
	"json-to-go/jsonparser"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	type args struct {
		jsonStr string
		config  *Config
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "基础对象解析",
			args: args{
				jsonStr: `{
  "k1": "v1",
  "k2": true,
  "k3": 3
}`,
				config: &Config{},
			},
			want: `type AutoGenerated struct {
	K1 string |json:"k1"|
	K2 bool   |json:"k2"|
	K3 int    |json:"k3"|
}`,
			wantErr: false,
		},
		{
			name: "数组类型解析",
			args: args{
				jsonStr: `{
  "k1": "v1",
  "k2": true,
  "k3": 3,
  "array1": [
    1,
    2
  ],
  "array2": [
    {
      "a": "b"
    }
  ],
  "array3": [
    [
      1,
      2
    ]
  ],
  "array4": [
    [
      {
        "a": "b"
      }
    ]
  ]
}`,
				config: &Config{},
			},
			want: `type AutoGenerated struct {
	K1     string     |json:"k1"|
	K2     bool       |json:"k2"|
	K3     int        |json:"k3"|
	Array1 []int      |json:"array1"|
	Array2 []Array2   |json:"array2"|
	Array3 [][]int    |json:"array3"|
	Array4 [][]Array4 |json:"array4"|
}

type Array2 struct {
	A string |json:"a"|
}

type Array4 struct {
	A string |json:"a"|
}`,
			wantErr: false,
		},
		{
			name: "测试属性的命名，支持中文，处理特殊字符",
			args: args{
				jsonStr: `{
  "地址": "山东枣庄",
  "地质": "中文解析后重名了，自动数字编号",
  "中文++))": "自动忽略不识别的字符",
  "id": "golint代码检查优化",
  "user_name": "",
  "3只松鼠": "",
  "doc_url": "http://localhost",
  "docUrl": "http://localhost",
  "curl": ""
}`,
				config: &Config{},
			},
			want: `type AutoGenerated struct {
	Dz       string |json:"地址"|
	Dz1      string |json:"地质"|
	Zw       string |json:"中文++))"|
	ID       string |json:"id"|
	UserName string |json:"user_name"|
	Threezss string |json:"3只松鼠"|
	DocURL   string |json:"doc_url"|
	DocURL1  string |json:"docUrl"|
	Curl     string |json:"curl"|
}`,
			wantErr: false,
		},
		{
			name: "测试属性类型推断",
			args: args{
				jsonStr: `{
  "float": 1.15,
  "arrayFloat": [
    1,
    2.1
  ],
  "arrayObjFloat1": [
    {
      "cost": 10
    },
    {
      "cost": 10.2
    }
  ],
  "arrayObjFloat2": [
    [
      {
        "cost": 10
      },
      {
        "cost": 10.2
      }
    ]
  ]
}`,
				config: &Config{},
			},
			want: `type AutoGenerated struct {
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
}`,
			wantErr: false,
		},
		{
			name: "测试空对象处理",
			args: args{
				jsonStr: `{}`,
				config:  &Config{},
			},
			want: `type AutoGenerated struct {
}`,
			wantErr: false,
		},
		{
			name: "测试属性为空对象处理",
			args: args{
				jsonStr: `{
  "a": {}
}`,
				config: &Config{},
			},
			want: `type AutoGenerated struct {
	A A |json:"a"|
}

type A struct {
}`,
			wantErr: false,
		},
		{
			name: "测试空值，空数组的处理",
			args: args{
				jsonStr: `{
  "array": [
    {
      "options": [],
      "other": null
    }
  ],
  "obj": {
    "options": [],
    "other": null
  },
  "options": [],
  "other": null
}`,
				config: &Config{},
			},
			want: `type AutoGenerated struct {
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
}`,
			wantErr: false,
		},
		{
			name: "支持入参是数组",
			args: args{
				jsonStr: `[
  {
    "id": 1,
    "name": "",
    "location": {
      "area": "",
      "province": ""
    }
  },
  {
    "id": 1,
    "age": 1,
    "location": {
      "city": ""
    }
  }
]`,
				config: &Config{},
			},
			want: `type AutoGenerated struct {
	ID       int      |json:"id"|
	Name     string   |json:"name"|
	Location Location |json:"location"|
	Age      int      |json:"age"|
}

type Location struct {
	Area     string |json:"area"|
	Province string |json:"province"|
	City     string |json:"city"|
}`,
			wantErr: false,
		},
		{
			name: "测试数组属性合并",
			args: args{
				jsonStr: `{
  "a": [
    {
      "a1": {
        "a2": [
          {
            "a3": ""
          },
          {
            "b3": ""
          }
        ]
      }
    },
    {
      "a1": {
        "a2": [
          {
            "a3": ""
          },
          {
            "c3": ""
          }
        ],
        "d2": "3"
      },
      "b1": {
        "b2": "3"
      }
    }
  ],
  "b": 1
}`,
				config: &Config{
					NestFlag: true,
				},
			},
			want: `type AutoGenerated struct {
	A []struct {
		A1 struct {
			A2 []struct {
				A3 string |json:"a3"|
				B3 string |json:"b3"|
				C3 string |json:"c3"|
			} |json:"a2"|
			D2 string |json:"d2"|
		} |json:"a1"|
		B1 struct {
			B2 string |json:"b2"|
		} |json:"b1"|
	} |json:"a"|
	B int |json:"b"|
}`,
			wantErr: false,
		},
		{
			name: "测试数组合并，同一个属性出现多个类型时，依赖类型判断来推断",
			args: args{
				jsonStr: `{
  "questionArray": [
    {
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
      "options": [
        {
          "label": "",
          "value": ""
        }
      ],
      "other": "",
      "required": false
    }
  ]
}`,
				config: &Config{},
			},
			want: `type AutoGenerated struct {
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
}`,
			wantErr: false,
		},
		{
			name: "测试使用指针",
			args: args{
				jsonStr: `{
  "location": {
    "province": ""
  },
  "location1": [
    {
      "province": ""
    }
  ],
  "location2": [
    [
      {
        "province": ""
      }
    ]
  ]
}`,
				config: &Config{
					PointerFlag: true,
				},
			},
			want: `type AutoGenerated struct {
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
}`,
			wantErr: false,
		},
		{
			name: "测试注释",
			args: args{
				jsonStr: `{
  // 1.支持单行注释提取，可在上一行或行尾
  "k1": "v1",
  "k2": "v2", // 行尾注释
  "k3": {
    "k3_1": {
      "k3_1_1": "" // 递归解析注释
    }
  },
  "k4": [{ // 这个注释会跟着k41
    "k41": ""
  }, {
    "k41": "" // 注释合并
  }]
}`,
				config: &Config{
					Comment: Comment1,
				},
			},
			want: `type AutoGenerated struct {
	// 1.支持单行注释提取，可在上一行或行尾
	K1 string |json:"k1"|
	// 行尾注释
	K2 string |json:"k2"|
	K3 K3     |json:"k3"|
	K4 []K4   |json:"k4"|
}

type K3 struct {
	K31 K31 |json:"k3_1"|
}

type K31 struct {
	// 递归解析注释
	K311 string |json:"k3_1_1"|
}

type K4 struct {
	// 这个注释会跟着k41
	K41 string |json:"k41"|
}`,
			wantErr: false,
		},
		{
			name: "测试注释中带双引号",
			args: args{
				jsonStr: `{
  "a": "", // 注释 ""
  "b": "", // 注释 ""
  "c": "", // 注释 ""
}`,
				config: &Config{
					Comment: Comment2,
				},
			},
			want: `type AutoGenerated struct {
	A string |json:"a"| // 注释 ""
	B string |json:"b"| // 注释 ""
	C string |json:"c"| // 注释 ""
}`,
			wantErr: false,
		},
		{
			name: "测试对象的复杂注释",
			args: args{
				jsonStr: `{
  // k1的注释
  // k1的注释2
  // k1的注释3
  "k1": "v1", // k1的注释4
  // k2的注释
  // k2的注释2
  // k2的注释3
  "k2": 2,
  // k3的注释
  "k3": 3 // k3的注释2
  // 这个注释去掉
  // 这个注释去掉
}`,
				config: &Config{
					Comment: Comment1,
				},
			},
			want: `type AutoGenerated struct {
	// k1的注释
	// k1的注释2
	// k1的注释3
	// k1的注释4
	K1 string |json:"k1"|
	// k2的注释
	// k2的注释2
	// k2的注释3
	K2 int |json:"k2"|
	// k3的注释
	// k3的注释2
	K3 int |json:"k3"|
}`,
			wantErr: false,
		},
		{
			name: "测试数组的复杂注释",
			args: args{
				jsonStr: `{
  "k1": [ // k1的注释
    // 这个注释去掉
    1 // 这个注释去掉
    // 这个注释去掉
    , // 这个注释去掉
    // 这个注释去掉
    2 // 这个注释去掉
    // 这个注释去掉
    // 这个注释去掉
  ],
  // k2的注释
  "k2": [ // 这个注释去掉
    // 这个注释去掉
    [ // 这个注释去掉
      1 // 这个注释去掉
      // 这个注释去掉
      , // 这个注释去掉
      // 这个注释去掉
      2 // 这个注释去掉
      // 这个注释去掉
    ] // 这个注释去掉
    // 这个注释去掉
  ]
}`,
				config: &Config{
					Comment: Comment2,
				},
			},
			want: `type AutoGenerated struct {
	K1 []int   |json:"k1"| // k1的注释
	K2 [][]int |json:"k2"| // k2的注释
}`,
			wantErr: false,
		},
		{
			name: "web使用的测试case",
			args: args{
				jsonStr: `{
  // 支持中文key
  "地址": "",
  "doc_url": "http://localhost", // golint命名优化
  // 重名会在末尾进行递增编号
  "docUrl": "http://localhost",
  // 整形数字，小于等于int32设置为int，大于int32设置为int64
  "int1": 1,
  "int2": 3000000000,
  // 浮点型数字，全部设置为float64
  "float": 1.15,
  // 如果是数组，会合并对象的属性和属性类型
  "a": [
    [{
      // 这个a1对象属性不全，会合并数组内其他a1对象的属性
      "a1": {
        "a2": [{
          "a3": "123" // 类型不同，判断为interface{}
        }, {
          "b3": ""
        }]
      }
    }, {
      "a1": {
        "a2": [{
          "a3": 123 // 类型不同，判断为interface{}
        }, {
          "c3": ""
        }],
        "b2": ""
      },
      "b1": {
        "b2": ""
      }
    }]
  ]
}`,
				config: &Config{
					Comment: Comment1,
				},
			},
			want: `type AutoGenerated struct {
	// 支持中文key
	Dz string |json:"地址"|
	// golint命名优化
	DocURL string |json:"doc_url"|
	// 重名会在末尾进行递增编号
	DocURL1 string |json:"docUrl"|
	// 整形数字，小于等于int32设置为int，大于int32设置为int64
	Int1 int   |json:"int1"|
	Int2 int64 |json:"int2"|
	// 浮点型数字，全部设置为float64
	Float float64 |json:"float"|
	// 如果是数组，会合并对象的属性和属性类型
	A [][]A |json:"a"|
}

type A struct {
	// 这个a1对象属性不全，会合并数组内其他a1对象的属性
	A1 A1 |json:"a1"|
	B1 B1 |json:"b1"|
}

type A1 struct {
	A2 []A2   |json:"a2"|
	B2 string |json:"b2"|
}

type A2 struct {
	// 类型不同，判断为interface{}
	A3 interface{} |json:"a3"|
	B3 string      |json:"b3"|
	C3 string      |json:"c3"|
}

type B1 struct {
	B2 string |json:"b2"|
}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want = strings.ReplaceAll(tt.want, "|", "`")
			got, err := Generate(tt.args.jsonStr, tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Generate() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func Test_getJSONType(t *testing.T) {
	type args struct {
		value []byte
		t     jsonparser.ValueType
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				value: []byte("10.5"),
				t:     jsonparser.Number,
			},
			want: TypeFloat64,
		},
		{
			args: args{
				value: []byte("10.0"),
				t:     jsonparser.Number,
			},
			want: TypeFloat64,
		},
		{
			args: args{
				value: []byte("0.0"),
				t:     jsonparser.Number,
			},
			want: TypeFloat64,
		},
		{
			args: args{
				value: []byte("100"),
				t:     jsonparser.Number,
			},
			want: TypeInt,
		},
		{
			args: args{
				value: []byte("3000000000"),
				t:     jsonparser.Number,
			},
			want: TypeInt64,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getJSONType(tt.args.value, tt.args.t); got != tt.want {
				t.Errorf("getJSONType() = %v, want %v", got, tt.want)
			}
		})
	}
}