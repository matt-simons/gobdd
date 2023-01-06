package models

import (
	"context"
	"errors"

	messages "github.com/cucumber/messages/go/v21"
)

type Feature struct {
	Location    *messages.Location       `json:"location"`
	Tags        []*messages.Tag          `json:"tags"`
	Language    string                   `json:"language"`
	Keyword     string                   `json:"keyword"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Children    []*messages.FeatureChild `json:"children"`

	//Rules []*Rule
	Scenarios []*Scenario // or Scenario Outline
}

func NewFeature(featureDoc *messages.Feature, scheme *Scheme) (*Feature, error) {
	f := &Feature{}
	var rules []*messages.Rule
	var backgrounds []*messages.Background
	var scenarios []*messages.Scenario

	for _, fc := range featureDoc.Children {
		if fc.Rule != nil {
			rules = append(rules, fc.Rule)
		}
		if fc.Background != nil {
			backgrounds = append(backgrounds, fc.Background)
		}
		if fc.Scenario != nil {
			scenarios = append(scenarios, fc.Scenario)
		}
	}

	//for _, RuleDoc := range rules {
	//	s, err := NewRule(ruleDoc, scheme)
	//	if err != nil {
	//		return f, err
	//	}
	//	f.Rules = append(f.Rules, s)
	//}

	if len(backgrounds) > 1 {
		return f, errors.New("a feature can only have one background")
	}

	var backgroundDoc *messages.Background
	if len(backgrounds) == 1 {
		backgroundDoc = backgrounds[0]
	}

	for _, scenarioDoc := range scenarios {
		s, err := NewScenario(backgroundDoc, scenarioDoc, scheme)
		if err != nil {
			return f, err
		}
		f.Scenarios = append(f.Scenarios, s)
	}
	return f, nil
}

// Add future parallel options
func (f *Feature) Run(ctx context.Context) {
	for _, scenario := range f.Scenarios {
		scenario.Run(ctx)
	}
}
