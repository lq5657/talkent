package role

type RoleType string

const (
	RoleTypeStructuredExpression RoleType = "structured_expression"
)

type Role struct {
	Description string   `json:"description"`
	Scenario    string   `json:"scenario"`
	Type        RoleType `json:"type"`
}

type TrainingGoal struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Dimension struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
