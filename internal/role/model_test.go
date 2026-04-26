package role

import (
	"encoding/json"
	"testing"
)

func TestRoleJSONSerialization(t *testing.T) {
	r := Role{
		Description: "模拟面试官",
		Scenario:    "技术面试",
		Type:        RoleTypeStructuredExpression,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal role: %v", err)
	}

	var got Role
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal role: %v", err)
	}

	if got.Description != r.Description {
		t.Errorf("description: got %q, want %q", got.Description, r.Description)
	}
	if got.Scenario != r.Scenario {
		t.Errorf("scenario: got %q, want %q", got.Scenario, r.Scenario)
	}
	if got.Type != r.Type {
		t.Errorf("type: got %q, want %q", got.Type, r.Type)
	}
}

func TestTrainingGoalJSONSerialization(t *testing.T) {
	g := TrainingGoal{Name: "逻辑条理性", Description: "观点是否有清晰的逻辑链条"}
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal goal: %v", err)
	}

	var got TrainingGoal
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal goal: %v", err)
	}
	if got.Name != g.Name {
		t.Errorf("name: got %q, want %q", got.Name, g.Name)
	}
}

func TestDimensionJSONSerialization(t *testing.T) {
	d := Dimension{Name: "论点清晰度", Description: "核心论点是否明确"}
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal dimension: %v", err)
	}

	var got Dimension
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal dimension: %v", err)
	}
	if got.Name != d.Name {
		t.Errorf("name: got %q, want %q", got.Name, d.Name)
	}
}

func TestMatchTemplate_Hit(t *testing.T) {
	tmpl, ok := MatchTemplate("我想模拟技术面试")
	if !ok {
		t.Fatal("expected template match for interview keyword")
	}
	if tmpl.Type != RoleTypeStructuredExpression {
		t.Errorf("type: got %q, want %q", tmpl.Type, RoleTypeStructuredExpression)
	}
	if len(tmpl.DefaultGoals) == 0 {
		t.Error("expected non-empty default goals")
	}
}

func TestMatchTemplate_Miss(t *testing.T) {
	_, ok := MatchTemplate("我想学习烹饪")
	if ok {
		t.Error("expected no template match for cooking keyword")
	}
}

func TestMatchTemplate_MultipleKeywords(t *testing.T) {
	keywords := []string{"演讲", "辩论", "汇报", "说服", "表达"}
	for _, kw := range keywords {
		_, ok := MatchTemplate(kw)
		if !ok {
			t.Errorf("expected match for keyword %q", kw)
		}
	}
}

func TestDimensionsForType_Valid(t *testing.T) {
	dims, ok := DimensionsForType(RoleTypeStructuredExpression)
	if !ok {
		t.Fatal("expected dimensions for structured_expression")
	}
	if len(dims) != 5 {
		t.Errorf("expected 5 dimensions, got %d", len(dims))
	}
	expected := []string{"论点清晰度", "论证结构", "口头禅检测", "回应度", "改进建议"}
	for i, d := range dims {
		if d.Name != expected[i] {
			t.Errorf("dim[%d]: got %q, want %q", i, d.Name, expected[i])
		}
	}
}

func TestDimensionsForType_Invalid(t *testing.T) {
	_, ok := DimensionsForType(RoleType("nonexistent"))
	if ok {
		t.Error("expected no dimensions for unknown type")
	}
}
