package gherkin

type Dialect struct {
	Language string
	Name     string
	Native   string
	Keywords map[string][]string
}

func (g *Dialect) FeatureKeywords() []string {
	return g.Keywords["feature"]
}

func (g *Dialect) RuleKeywords() []string {
	return g.Keywords["rule"]
}

func (g *Dialect) ScenarioKeywords() []string {
	return g.Keywords["scenario"]
}

func (g *Dialect) StepKeywords() []string {
	result := g.Keywords["given"]
	result = append(result, g.Keywords["when"]...)
	result = append(result, g.Keywords["then"]...)
	result = append(result, g.Keywords["and"]...)
	result = append(result, g.Keywords["but"]...)
	return result
}

func (g *Dialect) BackgroundKeywords() []string {
	return g.Keywords["background"]
}

func (g *Dialect) ScenarioOutlineKeywords() []string {
	return g.Keywords["scenarioOutline"]
}

func (g *Dialect) ExamplesKeywords() []string {
	return g.Keywords["examples"]
}

type DialectProvider interface {
	GetDialect(language string) *Dialect
}

type gherkinDialectMap map[string]*Dialect

func (g gherkinDialectMap) GetDialect(language string) *Dialect {
	return g[language]
}
