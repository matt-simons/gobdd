package gobdd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	gherkin "github.com/cucumber/gherkin/go/v26"
	msgs "github.com/cucumber/messages/go/v21"
)

// Suite holds all the information about the suite (options, steps to execute etc)
type Suite struct {
	steps          []stepDef
	options        SuiteOptions
	parameterTypes map[string][]string
}

// SuiteOptions holds all the information about how the suite or features/steps should be configured
type SuiteOptions struct {
	features       []string
	ignoreTags     []string
	tags           []string
	beforeScenario []func(ctx context.Context)
	afterScenario  []func(ctx context.Context)
	beforeStep     []func(ctx context.Context)
	afterStep      []func(ctx context.Context)
	runInParallel  bool
}

// WithFeaturesFS configures a filesystem and a path (glob pattern) where features can be found.
func WithFeaturesFS(path string) func(*SuiteOptions) {
	return func(options *SuiteOptions) {
		features, _ := fs.Glob(os.DirFS("."), path)
		options.features = features
	}
}

// NewSuiteOptions creates a new suite configuration with default values
func NewSuiteOptions() SuiteOptions {
	return SuiteOptions{
		//featureSource:  pathFeatureSource("features/*.feature"),
		ignoreTags:     []string{},
		tags:           []string{},
		beforeScenario: []func(ctx context.Context){},
		afterScenario:  []func(ctx context.Context){},
		beforeStep:     []func(ctx context.Context){},
		afterStep:      []func(ctx context.Context){},
	}
}

// RunInParallel runs tests in parallel
func RunInParallel() func(*SuiteOptions) {
	return func(options *SuiteOptions) {
		options.runInParallel = true
	}
}

// WithFeaturesPath configures a pattern (regexp) where feature can be found.
// The default value is "features/*.feature"
func WithFeaturesPath(path []string) func(*SuiteOptions) {
	return func(options *SuiteOptions) {
		options.features = path
	}
}

// WithTags configures which tags should be skipped while executing a suite
// Every tag has to start with @
func WithTags(tags ...string) func(*SuiteOptions) {
	return func(options *SuiteOptions) {
		options.tags = tags
	}
}

// WithBeforeScenario configures functions that should be executed before every scenario
func WithBeforeScenario(f func(ctx context.Context)) func(*SuiteOptions) {
	return func(options *SuiteOptions) {
		options.beforeScenario = append(options.beforeScenario, f)
	}
}

// WithAfterScenario configures functions that should be executed after every scenario
func WithAfterScenario(f func(ctx context.Context)) func(*SuiteOptions) {
	return func(options *SuiteOptions) {
		options.afterScenario = append(options.afterScenario, f)
	}
}

// WithBeforeStep configures functions that should be executed before every step
func WithBeforeStep(f func(ctx context.Context)) func(*SuiteOptions) {
	return func(options *SuiteOptions) {
		options.beforeStep = append(options.beforeStep, f)
	}
}

// WithAfterStep configures functions that should be executed after every step
func WithAfterStep(f func(ctx context.Context)) func(*SuiteOptions) {
	return func(options *SuiteOptions) {
		options.afterStep = append(options.afterStep, f)
	}
}

// WithIgnoredTags configures which tags should be skipped while executing a suite
// Every tag has to start with @ otherwise will be ignored
func WithIgnoredTags(tags ...string) func(*SuiteOptions) {
	return func(options *SuiteOptions) {
		options.ignoreTags = tags
	}
}

type stepDef struct {
	expr *regexp.Regexp
	f    interface{}
}

// Creates a new suites with given configuration and empty steps defined
func NewSuite(optionClosures ...func(*SuiteOptions)) *Suite {
	options := NewSuiteOptions()

	for i := 0; i < len(optionClosures); i++ {
		optionClosures[i](&options)
	}

	s := &Suite{
		steps:          []stepDef{},
		options:        options,
		parameterTypes: map[string][]string{},
	}

	s.AddParameterTypes(`{int}`, []string{`(\d)`})
	s.AddParameterTypes(`{float}`, []string{`([-+]?\d*\.?\d*)`})
	s.AddParameterTypes(`{word}`, []string{`([\d\w]+)`})
	s.AddParameterTypes(`{text}`, []string{`"([\d\w\-\s]+)"`, `'([\d\w\-\s]+)'`})

	return s
}

// AddParameterTypes adds a list of parameter types that will be used to simplify step definitions.
//
// The first argument is the parameter type and the second parameter is a list of regular expressions
// that should replace the parameter type.
//
//	s.AddParameterTypes(`{int}`, []string{`(\d)`})
//
// The regular expression should compile, otherwise will produce an error and stop executing.
func (s *Suite) AddParameterTypes(from string, to []string) {
	for _, to := range to {
		_, err := regexp.Compile(to)
		if err != nil {
			panic(fmt.Sprintf(`the regular expresion for key %s doesn't compile: %s`, from, to))
		}

		s.parameterTypes[from] = append(s.parameterTypes[from], to)
	}
}

// AddStep registers a step in the suite.
//
// The second parameter is the step function that gets executed
// when a step definition matches the provided regular expression.
//
// A step function can have any number of parameters (even zero),
// but it MUST accept a gobdd.StepTest and gobdd.Context as the first parameters (if there is any):
//
//	func myStepFunction(t gobdd.StepTest, ctx gobdd.Context, first int, second int) {
//	}
func (s *Suite) AddStep(expr string, step interface{}) {
	err := validateStepFunc(step)
	if err != nil {
		panic(fmt.Sprintf("the step function for step `%s` is incorrect: %w", expr, err))
	}

	exprs := s.applyParameterTypes(expr)

	for _, expr := range exprs {
		compiled := regexp.MustCompile(expr)
		s.steps = append(s.steps, stepDef{
			expr: compiled,
			f:    step,
		})
	}
}

func (s *Suite) applyParameterTypes(expr string) []string {
	exprs := []string{expr}

	for from, to := range s.parameterTypes {
		for _, t := range to {
			if strings.Contains(expr, from) {
				exprs = append(exprs, strings.ReplaceAll(expr, from, t))
			}
		}
	}

	return exprs
}

// AddRegexStep registers a step in the suite.
//
// The second parameter is the step function that gets executed
// when a step definition matches the provided regular expression.
//
// A step function can have any number of parameters (even zero),
// but it MUST accept a gobdd.StepTest and gobdd.Context as the first parameters (if there is any):
//
//	func myStepFunction(t gobdd.StepTest, ctx gobdd.Context, first int, second int) {
//	}
func (s *Suite) AddRegexStep(expr *regexp.Regexp, step interface{}) {
	err := validateStepFunc(step)
	if err != nil {
		panic(fmt.Sprintf("the step function is incorrect: %w", err))
	}

	s.steps = append(s.steps, stepDef{
		expr: expr,
		f:    step,
	})
}

// Executes the suite with given options and defined steps
func (s *Suite) Run() {

	for _, featurePath := range s.options.features {
		feature, err := os.Open(featurePath)

		doc, err := gherkin.ParseGherkinDocument(bufio.NewReader(feature), (&msgs.Incrementing{}).NewId)
		if err != nil {
			panic(fmt.Sprintf("error while loading document: %s\n", err))
		}
		defer feature.Close()

		if doc.Feature == nil {
			continue
		}

		s.runFeature(doc.Feature)
	}
}

func (s *Suite) runFeature(feature *msgs.Feature) {
	for _, tag := range feature.Tags {
		if contains(s.options.ignoreTags, tag.Name) {
			return
		}
	}

	for _, child := range feature.Children {
		if child.Scenario == nil {
			continue
		}

		if s.skipScenario(child.Scenario.Tags) {
			continue
		}

		// NewScenario(ctx, featureChild)
		s.runScenario(child.Scenario, child.Background)
	}
}

func (s *Suite) getOutlineStep(steps []*msgs.Step, examples []*msgs.Examples) []*msgs.Step {
	stepsList := make([][]*msgs.Step, len(steps))

	for i, outlineStep := range steps {
		for _, example := range examples {
			stepsList[i] = append(stepsList[i], s.stepsFromExamples(outlineStep, example)...)
		}
	}

	var newSteps []*msgs.Step

	if len(stepsList) == 0 {
		return newSteps
	}

	for ei := range examples {
		for ci := range examples[ei].TableBody {
			for si := range steps {
				newSteps = append(newSteps, stepsList[si][ci])
			}
		}
	}

	return newSteps
}

// generates steps
func (s *Suite) stepsFromExamples(sourceStep *msgs.Step, example *msgs.Examples) []*msgs.Step {
	steps := []*msgs.Step{}

	placeholders := example.TableHeader.Cells
	placeholdersValues := []string{}

	for _, placeholder := range placeholders {
		ph := "<" + placeholder.Value + ">"
		placeholdersValues = append(placeholdersValues, ph)
	}

	text := sourceStep.Text

	for _, row := range example.TableBody {
		// iterate over the cells and update the text
		stepText, expr := s.stepFromExample(text, row, placeholdersValues)

		// find step definition for the new step
		def, err := s.findStepDef(stepText)
		if err != nil {
			continue
		}

		// add the step to the list
		s.AddStep(expr, def.f)

		// clone a step
		step := &msgs.Step{
			Location: sourceStep.Location,
			Keyword:  sourceStep.Keyword,
			Text:     stepText,
			// TODO clone DocString and DocTable
		}

		steps = append(steps, step)
	}

	return steps
}

func (s *Suite) stepFromExample(stepName string, row *msgs.TableRow, placeholders []string) (string, string) {
	expr := stepName

	for i, ph := range placeholders {
		t := getRegexpForVar(row.Cells[i].Value)
		expr = strings.ReplaceAll(expr, ph, t)
		stepName = strings.ReplaceAll(stepName, ph, row.Cells[i].Value)
	}

	return stepName, expr
}

func (s *Suite) callBeforeScenarios(ctx context.Context) {
	for _, f := range s.options.beforeScenario {
		f(ctx)
	}
}

func (s *Suite) callAfterScenarios(ctx context.Context) {
	for _, f := range s.options.afterScenario {
		f(ctx)
	}
}

func (s *Suite) callBeforeSteps(ctx context.Context) {
	for _, f := range s.options.beforeStep {
		f(ctx)
	}
}

func (s *Suite) callAfterSteps(ctx context.Context) {
	for _, f := range s.options.afterStep {
		f(ctx)
	}
}

func (s *Suite) runScenario(scenario *msgs.Scenario, bkg *msgs.Background) {

	// TODO create kubernetes scenario
	// kubernetes scenario should incorporate runScenario, run, runStep, findStepDef and paramType

	ctx := context.Background()

	s.callBeforeScenarios(ctx)
	defer s.callAfterScenarios(ctx)

	if bkg != nil {
		for _, step := range bkg.Steps {
			s.runStep(ctx, step)
		}
	}

	if len(scenario.Examples) > 0 {
		steps := s.getOutlineStep(scenario.Steps, scenario.Examples)

		ctx := context.Background()
		for _, step := range steps {
			s.runStep(ctx, step)
		}
		return
	}

	for _, step := range scenario.Steps {
		s.runStep(ctx, step)
	}
}

func (s *Suite) runStep(ctx context.Context, step *msgs.Step) {
	def, err := s.findStepDef(step.Text)
	if err != nil {
		panic(fmt.Sprintf("cannot find step definition for step: %s%s", step.Keyword, step.Text))
	}

	params := def.expr.FindSubmatch([]byte(step.Text))[1:]

	s.callBeforeSteps(ctx)
	defer s.callAfterSteps(ctx)

	def.run(ctx, params)
}

func (def *stepDef) run(ctx context.Context, params [][]byte) {
	defer func() {
		if r := recover(); r != nil {
			// handle
		}
	}()

	d := reflect.ValueOf(def.f)
	if len(params)+1 != d.Type().NumIn() {
		panic(fmt.Sprintf("the step function %s accepts %d arguments but %d received", d.String(), d.Type().NumIn(), len(params)+1))
	}

	in := []reflect.Value{reflect.ValueOf(ctx)}

	for i, v := range params {
		if len(params) < i+1 {
			break
		}

		inType := d.Type().In(i + 1)
		paramType := paramType(v, inType)
		in = append(in, paramType)
	}

	d.Call(in)
}

func paramType(param []byte, inType reflect.Type) reflect.Value {
	paramType := reflect.ValueOf(param)
	if inType.Kind() == reflect.String {
		paramType = reflect.ValueOf(string(paramType.Interface().([]uint8)))
	}

	if inType.Kind() == reflect.Int {
		s := paramType.Interface().([]uint8)
		p, _ := strconv.Atoi(string(s))
		paramType = reflect.ValueOf(p)
	}

	if inType.Kind() == reflect.Float32 {
		s := paramType.Interface().([]uint8)
		p, _ := strconv.ParseFloat(string(s), 32)
		paramType = reflect.ValueOf(float32(p))
	}

	if inType.Kind() == reflect.Float64 {
		s := paramType.Interface().([]uint8)
		p, _ := strconv.ParseFloat(string(s), 32)
		paramType = reflect.ValueOf(p)
	}

	// add other types like boolean and StringOrInt

	return paramType
}

func (s *Suite) findStepDef(text string) (stepDef, error) {
	var sd stepDef

	found := 0
	matched := false

	for _, step := range s.steps {
		if !step.expr.MatchString(text) {
			continue
		}
		matched = true

		if l := len(step.expr.FindAll([]byte(text), -1)); l > found {
			found = l
			sd = step
		}
	}

	if !matched {
		return sd, errors.New("cannot find step definition")
	}

	return sd, nil
}

func (s *Suite) skipScenario(scenarioTags []*msgs.Tag) bool {
	for _, tag := range scenarioTags {
		if contains(s.options.ignoreTags, tag.Name) {
			return true
		}
	}

	if len(s.options.tags) == 0 {
		return false
	}

	for _, tag := range scenarioTags {
		if contains(s.options.tags, tag.Name) {
			return false
		}
	}

	return true
}

// contains tells whether a contains x.
func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}

	return false
}

func getRegexpForVar(v interface{}) string {
	s := v.(string)

	if _, err := strconv.Atoi(s); err == nil {
		return "(\\d+)"
	}

	if _, err := strconv.ParseFloat(s, 32); err == nil {
		return "([+-]?([0-9]*[.])?[0-9]+)"
	}

	return "(.*)"
}
