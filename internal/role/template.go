package role

import "strings"

type RoleTemplate struct {
	Type         RoleType
	Keywords     []string
	DefaultGoals []TrainingGoal
	DefaultDims  []Dimension
}

var builtInTemplates = []RoleTemplate{
	{
		Type: RoleTypeStructuredExpression,
		Keywords: []string{
			"面试", "汇报", "演讲", "辩论", "说服", "表达",
			"逻辑", "条理", "论证", "发言", "陈述", "报告",
		},
		DefaultGoals: []TrainingGoal{
			{Name: "逻辑条理性", Description: "观点是否有清晰的逻辑链条和层次结构"},
			{Name: "论证充分性", Description: "是否用事实、数据或案例支撑观点"},
			{Name: "重点突出", Description: "核心信息是否在有限时间内有效传达"},
			{Name: "语言精练", Description: "是否避免冗余和重复，用词精准"},
		},
		DefaultDims: []Dimension{
			{Name: "论点清晰度", Description: "核心论点是否明确、易懂"},
			{Name: "论证结构", Description: "论据是否有序支撑论点，逻辑链是否完整"},
			{Name: "口头禅检测", Description: "是否频繁使用填充词或无意义重复"},
			{Name: "回应度", Description: "是否有效回应对方的问题和观点"},
			{Name: "改进建议", Description: "针对本次对话的具体可操作改进建议"},
		},
	},
}

func MatchTemplate(desc string) (*RoleTemplate, bool) {
	lower := strings.ToLower(desc)
	for i := range builtInTemplates {
		t := &builtInTemplates[i]
		for _, kw := range t.Keywords {
			if strings.Contains(lower, kw) {
				return t, true
			}
		}
	}
	return nil, false
}

func DimensionsForType(rt RoleType) ([]Dimension, bool) {
	for _, t := range builtInTemplates {
		if t.Type == rt {
			return t.DefaultDims, true
		}
	}
	return nil, false
}
