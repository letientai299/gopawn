package gherkin

import (
	"gopawn/internal/msg"
	"strings"

	"github.com/iancoleman/strcase"
)

type AstBuilder interface {
	Builder
	GetGherkinDocument() *msg.GherkinDocument
}

type astBuilder struct {
	stack    []*astNode
	comments []*msg.GherkinDocument_Comment
}

func (t *astBuilder) Reset() {
	t.comments = []*msg.GherkinDocument_Comment{}
	t.stack = []*astNode{}
	t.push(newAstNode(RuleTypeNone))
}

func (t *astBuilder) GetGherkinDocument() *msg.GherkinDocument {
	res := t.currentNode().getSingle(RuleTypeGherkinDocument)
	if val, ok := res.(*msg.GherkinDocument); ok {
		return val
	}
	return nil
}

type astNode struct {
	ruleType RuleType
	subNodes map[RuleType][]interface{}
}

func (a *astNode) add(rt RuleType, obj interface{}) {
	a.subNodes[rt] = append(a.subNodes[rt], obj)
}

func (a *astNode) getSingle(rt RuleType) interface{} {
	if val, ok := a.subNodes[rt]; ok {
		for i := range val {
			return val[i]
		}
	}
	return nil
}

func (a *astNode) getItems(rt RuleType) []interface{} {
	var res []interface{}
	if val, ok := a.subNodes[rt]; ok {
		for i := range val {
			res = append(res, val[i])
		}
	}
	return res
}

func (a *astNode) getToken(tt TokenType) *Token {
	if val, ok := a.getSingle(tt.RuleType()).(*Token); ok {
		return val
	}
	return nil
}

func (a *astNode) getTokens(tt TokenType) []*Token {
	var items = a.getItems(tt.RuleType())
	var tokens []*Token
	for i := range items {
		if val, ok := items[i].(*Token); ok {
			tokens = append(tokens, val)
		}
	}
	return tokens
}

func (t *astBuilder) currentNode() *astNode {
	if len(t.stack) > 0 {
		return t.stack[len(t.stack)-1]
	}
	return nil
}

func newAstNode(rt RuleType) *astNode {
	return &astNode{
		ruleType: rt,
		subNodes: make(map[RuleType][]interface{}),
	}
}

func NewAstBuilder() AstBuilder {
	builder := new(astBuilder)
	builder.comments = []*msg.GherkinDocument_Comment{}
	builder.push(newAstNode(RuleTypeNone))
	return builder
}

func (t *astBuilder) push(n *astNode) {
	t.stack = append(t.stack, n)
}

func (t *astBuilder) pop() *astNode {
	x := t.stack[len(t.stack)-1]
	t.stack = t.stack[:len(t.stack)-1]
	return x
}

func (t *astBuilder) Build(tok *Token) (bool, error) {
	if tok.Type == TokenTypeComment {
		comment := &msg.GherkinDocument_Comment{
			Location: astLocation(tok),
			Text:     tok.Text,
		}
		t.comments = append(t.comments, comment)
	} else {
		t.currentNode().add(tok.Type.RuleType(), tok)
	}
	return true, nil
}
func (t *astBuilder) StartRule(r RuleType) (bool, error) {
	t.push(newAstNode(r))
	return true, nil
}
func (t *astBuilder) EndRule(r RuleType) (bool, error) {
	node := t.pop()
	transformedNode, err := t.transformNode(node)
	t.currentNode().add(node.ruleType, transformedNode)
	return true, err
}

func (t *astBuilder) transformNode(node *astNode) (interface{}, error) {
	switch node.ruleType {

	case RuleTypeStep:
		stepLine := node.getToken(TokenTypeStepLine)

		step := &msg.GherkinDocument_Feature_Step{
			Location: astLocation(stepLine),
			Keyword:  stepLine.Keyword,
			Text:     stepLine.Text,
		}
		dataTable := node.getSingle(RuleTypeDataTable)
		if dataTable != nil {
			step.Argument = &msg.GherkinDocument_Feature_Step_DataTable_{
				DataTable: dataTable.(*msg.GherkinDocument_Feature_Step_DataTable),
			}
		} else {
			docString := node.getSingle(RuleTypeDocString)
			if docString != nil {
				step.Argument = &msg.GherkinDocument_Feature_Step_DocString_{DocString: docString.(*msg.GherkinDocument_Feature_Step_DocString)}
			}
		}

		return step, nil

	case RuleTypeDocString:
		separatorToken := node.getToken(TokenTypeDocStringSeparator)
		lineTokens := node.getTokens(TokenTypeOther)
		var text string
		for i := range lineTokens {
			if i > 0 {
				text += "\n"
			}
			text += lineTokens[i].Text
		}
		ds := &msg.GherkinDocument_Feature_Step_DocString{
			Location:    astLocation(separatorToken),
			ContentType: separatorToken.Text,
			Content:     text,
			Delimiter:   separatorToken.Keyword,
		}
		return ds, nil

	case RuleTypeDataTable:
		rows, err := astTableRows(node)
		dt := &msg.GherkinDocument_Feature_Step_DataTable{
			Location: rows[0].Location,
			Rows:     rows,
		}
		return dt, err

	case RuleTypeBackground:
		backgroundLine := node.getToken(TokenTypeBackgroundLine)
		description, _ := node.getSingle(RuleTypeDescription).(string)
		bg := &msg.GherkinDocument_Feature_Background{
			Location:    astLocation(backgroundLine),
			Keyword:     backgroundLine.Keyword,
			Name:        backgroundLine.Text,
			Description: description,
			Steps:       astSteps(node),
		}
		return bg, nil

	case RuleTypeScenarioDefinition:
		tags := astTags(node)
		scenarioNode, _ := node.getSingle(RuleTypeScenario).(*astNode)

		scenarioLine := scenarioNode.getToken(TokenTypeScenarioLine)
		description, _ := scenarioNode.getSingle(RuleTypeDescription).(string)
		sc := &msg.GherkinDocument_Feature_Scenario{
			Tags:        tags,
			Location:    astLocation(scenarioLine),
			Keyword:     scenarioLine.Keyword,
			Name:        scenarioLine.Text,
			Description: description,
			Steps:       astSteps(scenarioNode),
			Examples:    astExamples(scenarioNode),
		}

		return sc, nil

	case RuleTypeExamplesDefinition:
		tags := astTags(node)
		examplesNode, _ := node.getSingle(RuleTypeExamples).(*astNode)
		examplesLine := examplesNode.getToken(TokenTypeExamplesLine)
		description, _ := examplesNode.getSingle(RuleTypeDescription).(string)
		examplesTable := examplesNode.getSingle(RuleTypeExamplesTable)

		// TODO: Is this mutation style ok?
		ex := &msg.GherkinDocument_Feature_Scenario_Examples{}
		ex.Tags = tags
		ex.Location = astLocation(examplesLine)
		ex.Keyword = examplesLine.Keyword
		ex.Name = examplesLine.Text
		ex.Description = description
		ex.TableHeader = nil
		ex.TableBody = nil
		if examplesTable != nil {
			allRows, _ := examplesTable.([]*msg.GherkinDocument_Feature_TableRow)
			ex.TableHeader = allRows[0]
			ex.TableBody = allRows[1:]
		}
		return ex, nil

	case RuleTypeExamplesTable:
		allRows, err := astTableRows(node)
		return allRows, err

	case RuleTypeDescription:
		lineTokens := node.getTokens(TokenTypeOther)
		// Trim trailing empty lines
		end := len(lineTokens)
		for end > 0 && strings.TrimSpace(lineTokens[end-1].Text) == "" {
			end--
		}
		var desc []string
		for i := range lineTokens[0:end] {
			desc = append(desc, lineTokens[i].Text)
		}
		return strings.Join(desc, "\n"), nil

	case RuleTypeProgram:
		header, ok := node.getSingle(RuleTypeProgramHeader).(*astNode)
		if !ok {
			return nil, nil
		}
		progLine := header.getToken(TokenTypeProgramLine)
		if progLine == nil {
			return nil, nil
		}

		description, _ := header.getSingle(RuleTypeDescription).(string)

		prog := &msg.GherkinDocument_Program{}
		prog.Location = astLocation(progLine)
		prog.Language = progLine.GherkinDialect
		prog.Keyword = progLine.Keyword
		prog.Name = strcase.ToSnake(progLine.Text)
		prog.Description = description
		return prog, nil

	case RuleTypeFeature:
		header, ok := node.getSingle(RuleTypeFeatureHeader).(*astNode)
		if !ok {
			return nil, nil
		}
		tags := astTags(header)
		featureLine := header.getToken(TokenTypeFeatureLine)
		if featureLine == nil {
			return nil, nil
		}

		var children []*msg.GherkinDocument_Feature_FeatureChild
		background, _ := node.getSingle(RuleTypeBackground).(*msg.GherkinDocument_Feature_Background)
		if background != nil {
			children = append(children, &msg.GherkinDocument_Feature_FeatureChild{
				Value: &msg.GherkinDocument_Feature_FeatureChild_Background{Background: background},
			})
		}
		scenarios := node.getItems(RuleTypeScenarioDefinition)
		for i := range scenarios {
			scenario := scenarios[i].(*msg.GherkinDocument_Feature_Scenario)
			children = append(children, &msg.GherkinDocument_Feature_FeatureChild{
				Value: &msg.GherkinDocument_Feature_FeatureChild_Scenario{Scenario: scenario},
			})
		}
		rules := node.getItems(RuleTypeRule)
		for i := range rules {
			rule := rules[i].(*msg.GherkinDocument_Feature_FeatureChild_Rule)
			children = append(children, &msg.GherkinDocument_Feature_FeatureChild{
				Value: &msg.GherkinDocument_Feature_FeatureChild_Rule_{
					Rule: rule,
				},
			})
		}

		description, _ := header.getSingle(RuleTypeDescription).(string)

		feat := &msg.GherkinDocument_Feature{}
		feat.Tags = tags
		feat.Location = astLocation(featureLine)
		feat.Language = featureLine.GherkinDialect
		feat.Keyword = featureLine.Keyword
		feat.Name = featureLine.Text
		feat.Description = description
		feat.Children = children
		return feat, nil

	case RuleTypeRule:
		header, ok := node.getSingle(RuleTypeRuleHeader).(*astNode)
		if !ok {
			return nil, nil
		}
		ruleLine := header.getToken(TokenTypeRuleLine)
		if ruleLine == nil {
			return nil, nil
		}

		var children []*msg.GherkinDocument_Feature_FeatureChild_RuleChild
		background, _ := node.getSingle(RuleTypeBackground).(*msg.GherkinDocument_Feature_Background)

		if background != nil {
			children = append(children, &msg.GherkinDocument_Feature_FeatureChild_RuleChild{
				Value: &msg.GherkinDocument_Feature_FeatureChild_RuleChild_Background{Background: background},
			})
		}
		scenarios := node.getItems(RuleTypeScenarioDefinition)
		for i := range scenarios {
			scenario := scenarios[i].(*msg.GherkinDocument_Feature_Scenario)
			children = append(children, &msg.GherkinDocument_Feature_FeatureChild_RuleChild{
				Value: &msg.GherkinDocument_Feature_FeatureChild_RuleChild_Scenario{Scenario: scenario},
			})
		}

		description, _ := header.getSingle(RuleTypeDescription).(string)

		rule := &msg.GherkinDocument_Feature_FeatureChild_Rule{}
		rule.Location = astLocation(ruleLine)
		rule.Keyword = ruleLine.Keyword
		rule.Name = ruleLine.Text
		rule.Description = description
		rule.Children = children
		return rule, nil

	case RuleTypeGherkinDocument:
		feature, _ := node.getSingle(RuleTypeFeature).(*msg.GherkinDocument_Feature)
		prog, _ := node.getSingle(RuleTypeProgram).(*msg.GherkinDocument_Program)

		doc := &msg.GherkinDocument{}
		if feature != nil {
			doc.Feature = feature
		}
		if prog != nil {
			doc.Program = prog
		}
		doc.Comments = t.comments
		return doc, nil
	}
	return node, nil
}

func astLocation(t *Token) *msg.Location {
	return &msg.Location{
		Line:   uint32(t.Location.Line),
		Column: uint32(t.Location.Column),
	}
}

func astTableRows(t *astNode) (rows []*msg.GherkinDocument_Feature_TableRow, err error) {
	rows = []*msg.GherkinDocument_Feature_TableRow{}
	tokens := t.getTokens(TokenTypeTableRow)
	for i := range tokens {
		row := &msg.GherkinDocument_Feature_TableRow{
			Location: astLocation(tokens[i]),
			Cells:    astTableCells(tokens[i]),
		}
		rows = append(rows, row)
	}
	err = ensureCellCount(rows)
	return
}

func ensureCellCount(rows []*msg.GherkinDocument_Feature_TableRow) error {
	if len(rows) <= 1 {
		return nil
	}
	cellCount := len(rows[0].Cells)
	for i := range rows {
		if cellCount != len(rows[i].Cells) {
			return &parseError{"inconsistent cell count within the table", &Location{
				Line:   int(rows[i].Location.Line),
				Column: int(rows[i].Location.Column),
			}}
		}
	}
	return nil
}

func astTableCells(t *Token) (cells []*msg.GherkinDocument_Feature_TableRow_TableCell) {
	cells = []*msg.GherkinDocument_Feature_TableRow_TableCell{}
	for i := range t.Items {
		item := t.Items[i]
		cell := &msg.GherkinDocument_Feature_TableRow_TableCell{}
		cell.Location = &msg.Location{
			Line:   uint32(t.Location.Line),
			Column: uint32(item.Column),
		}
		cell.Value = item.Text
		cells = append(cells, cell)
	}
	return
}

func astSteps(t *astNode) (steps []*msg.GherkinDocument_Feature_Step) {
	steps = []*msg.GherkinDocument_Feature_Step{}
	tokens := t.getItems(RuleTypeStep)
	for i := range tokens {
		step, _ := tokens[i].(*msg.GherkinDocument_Feature_Step)
		steps = append(steps, step)
	}
	return
}

func astExamples(t *astNode) (examples []*msg.GherkinDocument_Feature_Scenario_Examples) {
	examples = []*msg.GherkinDocument_Feature_Scenario_Examples{}
	tokens := t.getItems(RuleTypeExamplesDefinition)
	for i := range tokens {
		example, _ := tokens[i].(*msg.GherkinDocument_Feature_Scenario_Examples)
		examples = append(examples, example)
	}
	return
}

func astTags(node *astNode) (tags []*msg.GherkinDocument_Feature_Tag) {
	tags = []*msg.GherkinDocument_Feature_Tag{}
	tagsNode, ok := node.getSingle(RuleTypeTags).(*astNode)
	if !ok {
		return
	}
	tokens := tagsNode.getTokens(TokenTypeTagLine)
	for i := range tokens {
		token := tokens[i]
		for k := range token.Items {
			item := token.Items[k]
			tag := &msg.GherkinDocument_Feature_Tag{}
			tag.Location = &msg.Location{
				Line:   uint32(token.Location.Line),
				Column: uint32(item.Column),
			}
			tag.Name = item.Text
			tags = append(tags, tag)
		}
	}
	return
}
