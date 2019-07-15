package gherkin

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"gopawn/internal/msg"
	"strings"
	"unicode/utf8"
)

func Pickles(gherkinDocument msg.GherkinDocument, uri string, source string) []*msg.Pickle {
	pickles := make([]*msg.Pickle, 0)
	if gherkinDocument.Feature == nil {
		return pickles
	}
	language := gherkinDocument.Feature.Language

	pickles = compileFeature(pickles, *gherkinDocument.Feature, uri, language, source)
	return pickles
}

func compileFeature(pickles []*msg.Pickle, feature msg.GherkinDocument_Feature, uri string, language string, source string) []*msg.Pickle {
	backgroundSteps := make([]*msg.Pickle_PickleStep, 0)
	featureTags := feature.Tags
	for _, child := range feature.Children {
		switch t := child.Value.(type) {
		case *msg.GherkinDocument_Feature_FeatureChild_Background:
			backgroundSteps = append(backgroundSteps, pickleSteps(t.Background.Steps)...)
		case *msg.GherkinDocument_Feature_FeatureChild_Rule_:
			pickles = compileRule(pickles, child.GetRule(), featureTags, backgroundSteps, uri, language, source)
		case *msg.GherkinDocument_Feature_FeatureChild_Scenario:
			scenario := t.Scenario
			if len(scenario.GetExamples()) == 0 {
				pickles = compileScenario(pickles, backgroundSteps, scenario, featureTags, uri, language, source)
			} else {
				pickles = compileScenarioOutline(pickles, scenario, featureTags, backgroundSteps, uri, language, source)
			}
		default:
			panic(fmt.Sprintf("unexpected %T feature child", child))
		}
	}
	return pickles
}

func compileRule(pickles []*msg.Pickle, rule *msg.GherkinDocument_Feature_FeatureChild_Rule, tags []*msg.GherkinDocument_Feature_Tag, steps []*msg.Pickle_PickleStep, uri string, language string, source string) []*msg.Pickle {
	backgroundSteps := make([]*msg.Pickle_PickleStep, 0)
	backgroundSteps = append(backgroundSteps, steps...)

	for _, child := range rule.Children {
		switch t := child.Value.(type) {
		case *msg.GherkinDocument_Feature_FeatureChild_RuleChild_Background:
			backgroundSteps = append(backgroundSteps, pickleSteps(t.Background.Steps)...)
		case *msg.GherkinDocument_Feature_FeatureChild_RuleChild_Scenario:
			scenario := t.Scenario
			if len(scenario.GetExamples()) == 0 {
				pickles = compileScenario(pickles, backgroundSteps, scenario, tags, uri, language, source)
			} else {
				pickles = compileScenarioOutline(pickles, scenario, tags, backgroundSteps, uri, language, source)
			}
		default:
			panic(fmt.Sprintf("unexpected %T feature child", child))
		}
	}
	return pickles

}

func compileScenarioOutline(pickles []*msg.Pickle, scenario *msg.GherkinDocument_Feature_Scenario, featureTags []*msg.GherkinDocument_Feature_Tag, backgroundSteps []*msg.Pickle_PickleStep, uri string, language string, source string) []*msg.Pickle {
	for _, examples := range scenario.Examples {
		if examples.TableHeader == nil {
			continue
		}
		variableCells := examples.TableHeader.Cells
		for _, values := range examples.TableBody {
			valueCells := values.Cells
			tags := pickleTags(append(featureTags, append(scenario.Tags, examples.Tags...)...))

			pickleSteps := make([]*msg.Pickle_PickleStep, 0)

			// translate pickleSteps based on values
			for _, step := range scenario.Steps {
				text := step.Text
				for i, variableCell := range variableCells {
					text = strings.Replace(text, "<"+variableCell.Value+">", valueCells[i].Value, -1)
				}

				pickleStep := pickleStep(step, variableCells, valueCells)
				pickleStep.Locations = append(pickleStep.Locations, values.Location)
				pickleSteps = append(pickleSteps, pickleStep)
			}

			// translate pickle name
			name := scenario.Name
			for i, key := range variableCells {
				name = strings.Replace(name, "<"+key.Value+">", valueCells[i].Value, -1)
			}

			if len(pickleSteps) > 0 {
				pickleSteps = append(backgroundSteps, pickleSteps...)
			}

			locations := make([]*msg.Location, 0)
			locations = append(locations, scenario.Location)
			locations = append(locations, values.Location)

			pickles = append(pickles, &msg.Pickle{
				Id:        makeId(source, locations),
				Uri:       uri,
				Steps:     pickleSteps,
				Tags:      tags,
				Name:      name,
				Language:  language,
				Locations: locations,
			})
		}
	}
	return pickles
}

func makeId(source string, locations []*msg.Location) string {
	hash := sha1.New()
	hash.Write([]byte(source))
	for _, location := range locations {
		err := binary.Write(hash, binary.LittleEndian, location.Line)
		if err != nil {
			panic(err)
		}
		err = binary.Write(hash, binary.LittleEndian, location.Column)
		if err != nil {
			panic(err)
		}
	}
	bs := hash.Sum(nil)
	return hex.EncodeToString(bs)
}

func compileScenario(pickles []*msg.Pickle, backgroundSteps []*msg.Pickle_PickleStep, scenario *msg.GherkinDocument_Feature_Scenario, featureTags []*msg.GherkinDocument_Feature_Tag, uri string, language string, source string) []*msg.Pickle {
	steps := make([]*msg.Pickle_PickleStep, 0)
	if len(scenario.Steps) > 0 {
		steps = append(backgroundSteps, pickleSteps(scenario.Steps)...)
	}
	tags := pickleTags(append(featureTags, scenario.Tags...))
	locations := make([]*msg.Location, 0)
	locations = append(locations, scenario.Location)
	pickles = append(pickles, &msg.Pickle{
		Id:        makeId(source, locations),
		Uri:       uri,
		Steps:     steps,
		Tags:      tags,
		Name:      scenario.Name,
		Language:  language,
		Locations: locations,
	})
	return pickles
}

func pickleDataTable(table *msg.GherkinDocument_Feature_Step_DataTable, variableCells, valueCells []*msg.GherkinDocument_Feature_TableRow_TableCell) *msg.PickleStepArgument_PickleTable {
	pickleTableRows := make([]*msg.PickleStepArgument_PickleTable_PickleTableRow, len(table.Rows))
	for i, row := range table.Rows {
		pickleTableCells := make([]*msg.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell, len(row.Cells))
		for j, cell := range row.Cells {
			pickleTableCells[j] = &msg.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{
				Location: cell.Location,
				Value:    interpolate(cell.Value, variableCells, valueCells),
			}
		}
		pickleTableRows[i] = &msg.PickleStepArgument_PickleTable_PickleTableRow{Cells: pickleTableCells}
	}
	return &msg.PickleStepArgument_PickleTable{Rows: pickleTableRows}
}

func pickleDocString(docString *msg.GherkinDocument_Feature_Step_DocString, variableCells, valueCells []*msg.GherkinDocument_Feature_TableRow_TableCell) *msg.PickleStepArgument_PickleDocString {
	return &msg.PickleStepArgument_PickleDocString{
		Location:    docString.Location,
		ContentType: interpolate(docString.ContentType, variableCells, valueCells),
		Content:     interpolate(docString.Content, variableCells, valueCells),
	}
}

func pickleTags(tags []*msg.GherkinDocument_Feature_Tag) []*msg.Pickle_PickleTag {
	ptags := make([]*msg.Pickle_PickleTag, len(tags))
	for i, tag := range tags {
		ptags[i] = &msg.Pickle_PickleTag{
			Location: tag.Location,
			Name:     tag.Name,
		}
	}
	return ptags
}

func pickleSteps(steps []*msg.GherkinDocument_Feature_Step) []*msg.Pickle_PickleStep {
	pickleSteps := make([]*msg.Pickle_PickleStep, len(steps))
	for i, step := range steps {
		pickleStep := pickleStep(step, nil, nil)
		pickleSteps[i] = pickleStep
	}
	return pickleSteps
}

func pickleStep(step *msg.GherkinDocument_Feature_Step, variableCells, valueCells []*msg.GherkinDocument_Feature_TableRow_TableCell) *msg.Pickle_PickleStep {
	loc := &msg.Location{
		Line:   step.Location.Line,
		Column: step.Location.Column + uint32(utf8.RuneCountInString(step.Keyword)),
	}
	locations := make([]*msg.Location, 0)
	locations = append(locations, loc)
	pickleStep := &msg.Pickle_PickleStep{
		Text:      interpolate(step.Text, variableCells, valueCells),
		Locations: locations,
	}
	if step.GetDataTable() != nil {
		pickleStep.Argument = &msg.PickleStepArgument{
			Message: &msg.PickleStepArgument_DataTable{
				DataTable: pickleDataTable(step.GetDataTable(), variableCells, valueCells),
			},
		}
	}
	if step.GetDocString() != nil {
		pickleStep.Argument = &msg.PickleStepArgument{
			Message: &msg.PickleStepArgument_DocString{
				DocString: pickleDocString(step.GetDocString(), variableCells, valueCells),
			},
		}
	}
	return pickleStep
}

func interpolate(s string, variableCells, valueCells []*msg.GherkinDocument_Feature_TableRow_TableCell) string {
	if variableCells == nil || valueCells == nil {
		return s
	}

	for i, variableCell := range variableCells {
		s = strings.Replace(s, "<"+variableCell.Value+">", valueCells[i].Value, -1)
	}

	return s
}
