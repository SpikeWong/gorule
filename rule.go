package gorule

type Rule struct {
	name      string
	condition string
	action    func(interface{}) (interface{}, error)
}

// NewRule creates rule with trigger condition and action function to be
// executed when the condition is met.
func NewRule(name, condition string, action func(interface{}) (interface{}, error)) *Rule {
	return &Rule{name: name, condition: condition, action: action}
}

// Name returns the name of rule.
func (r *Rule) Name() string {
	return r.name
}

// Execute will execute action function with input.
func (r *Rule) Execute(input interface{}) (interface{}, error) {
	return r.action(input)
}
