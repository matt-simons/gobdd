package models

import (
	"context"

	messages "github.com/cucumber/messages/go/v21"
)

type Scenario struct {
	Location    *messages.Location   `json:"location"`
	Tags        []*messages.Tag      `json:"tags"`
	Keyword     string               `json:"keyword"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Background  *messages.Background `json:"background"`
	Steps       []*Step              `json:"steps"`
	//Examples    []*Examples        `json:"examples"`
}

func NewScenario(bkg *messages.Background, scn *messages.Scenario, scheme *Scheme) (*Scenario, error) {
	s := &Scenario{
		Location:   scn.Location,
		Tags:       scn.Tags,
		Keyword:    scn.Keyword,
		Name:       scn.Name,
		Background: bkg,
	}

	bkgSteps, err := GenerateSteps(bkg.Steps, scheme)
	if err != nil {
		return s, err
	}
	scnSteps, err := GenerateSteps(scn.Steps, scheme)
	if err != nil {
		return s, err
	}
	s.Steps = append(bkgSteps, scnSteps...)

	return s, nil
}

func (s *Scenario) Run(ctx context.Context) {
	// add to ctx
	// * Helper
	// * Register
	// * PodSessions
	// * PortForwarders
	// * out and errOut Writers
	for _, step := range s.Steps {
		step.Run(ctx)
		if step.Execution.Result != Passed {
			break
		}
	}
}

type Background struct {
	Location    *messages.Location `json:"location"`
	Keyword     string             `json:"keyword"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Steps       []*Step            `json:"steps"`
}
