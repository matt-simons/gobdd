package models

import (
	"context"
	"reflect"
	"time"

	messages "github.com/cucumber/messages/go/v21"
)

type Step struct {
	// Should these if templated by hydrated? yes
	Location    *messages.Location       `json:"location"`
	Keyword     string                   `json:"keyword"`
	KeywordType messages.StepKeywordType `json:"keywordType,omitempty"`
	Text        string                   `json:"text"`
	DocString   *messages.DocString      `json:"docString,omitempty"`
	DataTable   *messages.DataTable      `json:"dataTable,omitempty"`

	// Step Definition
	Func reflect.Value
	Args []reflect.Value

	// Step Result
	Execution StepExecution `json:"execution"`
}

type StepExecution struct {
	Result    Result
	StartTime time.Time
	EndTime   time.Time
	Err       error
}

type Result int

const (
	Passed Result = iota
	Failed
	Skipped
)

func (s *Step) Run(ctx context.Context) {
	// ctx is the scenario context
	// it contains an overall deadline or timeout for feature/scenario
	// it contains the registry/pod sessions/helper etc
	args := append([]reflect.Value{reflect.ValueOf(ctx)}, s.Args...)

	s.Execution.StartTime = time.Now()
	ret := s.Func.Call(args)
	s.Execution.EndTime = time.Now()

	if len(ret) != 1 {
		panic("steps should only return a single error or nil")
	}

	if ret[0].IsNil() {
		s.Execution.Result = Passed
		return
	}

	r := ret[0].Interface()
	if err, ok := r.(error); ok {
		s.Execution.Result = Failed
		s.Execution.Err = err
		return
	}
	panic("steps should only return a single error or nil")
}

func NewStep(stepDoc *messages.Step, scheme Scheme) (*Step, error) {
	s := &Step{
		Location:    stepDoc.Location,
		Keyword:     stepDoc.Keyword,
		KeywordType: stepDoc.KeywordType,
		Text:        stepDoc.Text,
		DocString:   stepDoc.DocString,
		DataTable:   stepDoc.DataTable,
	}

	err := scheme.StepDefFor(s)
	if err != nil {
		return nil, err
	}

	return s, nil
}
