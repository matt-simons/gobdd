package models

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	messages "github.com/cucumber/messages/go/v21"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var basicGoodStep = StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function: func(ctx context.Context, s string) error {
		if len(s) > 1 {
			return nil
		}
		return fmt.Errorf("small string")
	},
}

var basicStepWithoutContext = StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function: func(s string) error {
		if len(s) > 1 {
			return nil
		}
		return fmt.Errorf("small string")
	},
}

var basicStepWithoutArgs = StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function: func() error {
		return nil
	},
}

var basicStepWithoutFunc = StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function:   "notafunction",
}

var basicStepTooFewArgs = StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function: func(ctx context.Context) error {
		return nil
	},
}

var basicStepTooManyArgs = StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function: func(ctx context.Context, s, b string) error {
		return nil
	},
}

var basicGoodDocStringStep = StepDefinition{
	Expression: regexp.MustCompile("find all (.*) in:"),
	Function: func(ctx context.Context, s string, doc *messages.DocString) error {
		return nil
	},
}

var basicTooManyDocStringStep = StepDefinition{
	Expression: regexp.MustCompile("find all (.*) in:"),
	Function: func(ctx context.Context, s string, doc *messages.DocString, doc2 *messages.DocString) error {
		return nil
	},
}

var basicGoodDataTableStep = StepDefinition{
	Expression: regexp.MustCompile("find all (.*) in:"),
	Function: func(ctx context.Context, s string, doc *messages.DataTable) error {
		return nil
	},
}

var basicTooManyArgStep = StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function: func(ctx context.Context, s string, doc *messages.DataTable) error {
		return nil
	},
}

var _ = Describe("Registering Steps", func() {

	Context("Registering Basic Steps", func() {

		It("should register a basic good step", func() {
			scheme := Scheme{}
			Expect(scheme.Register(basicGoodStep)).Should(Succeed())
		})

		It("should not register a basic step without a context", func() {
			scheme := Scheme{}
			Expect(scheme.Register(basicStepWithoutContext)).Should(MatchError(ErrMustHaveContext))
		})

		It("should not register a basic step without any arguments", func() {
			scheme := Scheme{}
			Expect(scheme.Register(basicStepWithoutArgs)).Should(MatchError(ErrMustHaveContext))
		})

		It("should not register a basic step without a function", func() {
			scheme := Scheme{}
			Expect(scheme.Register(basicStepWithoutFunc)).Should(MatchError(ErrStepDefinitionMustHaveFunc))
		})

		It("should not register a basic step which has too few args for the regular expression", func() {
			scheme := Scheme{}
			Expect(scheme.Register(basicStepTooFewArgs)).Should(MatchError(ErrTooFewArguments))
		})

		It("should not register a basic step which has too many args for the regular expression", func() {
			scheme := Scheme{}
			Expect(scheme.Register(basicStepTooManyArgs)).Should(MatchError(ErrTooManyArguments))
		})
	})

	Context("Registering Steps with DocString or DataTable Arguments", func() {

		It("should register a good step with DocString", func() {
			scheme := Scheme{}
			Expect(scheme.Register(basicGoodDocStringStep)).Should(Succeed())
		})

		It("should register a good step with DataTable", func() {
			scheme := Scheme{}
			Expect(scheme.Register(basicGoodDataTableStep)).Should(Succeed())
		})

		It("should not register a step with two DocString", func() {
			scheme := Scheme{}
			Expect(scheme.Register(basicTooManyDocStringStep)).Should(MatchError(ErrTooManyArguments))
		})
	})
})

var _ = Describe("Hydrating a Step", func() {

	Context("Applying a Step Definition to a Step", func() {
		DescribeTable("Applying a matching step definition to a step",
			func(text string, arg interface{}, f interface{}) {
				var step = &Step{
					Text: text,
				}
				var stepDef = StepDefinition{
					Expression: regexp.MustCompile("a (.*)"),
					Function:   f,
				}
				scheme := Scheme{}
				Expect(scheme.Register(stepDef)).Should(Succeed())
				Expect(scheme.StepDefFor(step)).Should(Succeed())
				Expect(step.Func).Should(Equal(reflect.ValueOf(f)))
				Expect(step.Args).Should(HaveLen(1))
				Expect(step.Args[0].Interface()).Should(Equal(arg))
			},
			Entry("When arg is a string", "a word", "word", func(ctx context.Context, a string) error { return nil }),
			Entry("When arg is a int", "a 47", 47, func(ctx context.Context, a int) error { return nil }),
			Entry("When arg is a int8", "a 8", int8(8), func(ctx context.Context, a int8) error { return nil }),
			Entry("When arg is a int16", "a 16", int16(16), func(ctx context.Context, a int16) error { return nil }),
			Entry("When arg is a int32", "a 32", int32(32), func(ctx context.Context, a int32) error { return nil }),
			Entry("When arg is a int64", "a 64", int64(64), func(ctx context.Context, a int64) error { return nil }),
			Entry("When arg is a float32", "a 3.2", float32(3.2), func(ctx context.Context, a float32) error { return nil }),
			Entry("When arg is a float64", "a 6.4", float64(6.4), func(ctx context.Context, a float64) error { return nil }),
			Entry("When arg is a []byte", "a bytes", []byte("bytes"), func(ctx context.Context, a []byte) error { return nil }),
		)

		DescribeTable("Applying a matching step definition with DocString or DataTable to a step",
			func(arg interface{}, f interface{}) {
				var step = &Step{
					Text: "a blah",
				}
				if ds, ok := arg.(*messages.DocString); ok {
					step.DocString = ds
				}
				if dt, ok := arg.(*messages.DataTable); ok {
					step.DataTable = dt
				}
				var stepDef = StepDefinition{
					Expression: regexp.MustCompile("a (.*)"),
					Function:   f,
				}
				scheme := Scheme{}
				Expect(scheme.Register(stepDef)).Should(Succeed())
				Expect(scheme.StepDefFor(step)).Should(Succeed())
				Expect(step.Func).Should(Equal(reflect.ValueOf(f)))
				Expect(step.Args).Should(HaveLen(2))
				Expect(step.Args[1].Interface()).Should(Equal(arg))
			},
			Entry("When DocString", &messages.DocString{Content: "helloworld"}, func(ctx context.Context, s string, doc *messages.DocString) error { return nil }),
			Entry("When DataTable", &messages.DataTable{}, func(ctx context.Context, s string, doc *messages.DataTable) error { return nil }),
		)

	})
})

func TestScheme(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Books Suite")
}
