package gobdd

import (
	"context"
	"testing"

	"github.com/go-bdd/gobdd"
)

func TestValidateStepFunc(t *testing.T) {
	testCases := map[string]interface{}{
		"function without arguments":           func() {},
		"function with 1 argument":             func(StepTest) {},
		"function with invalid first argument": func(int, Context) {},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			if err := validateStepFunc(testCase); err == nil {
				t.Errorf("the test should fail for the function")
			}
		})
	}
}

// Used for context package backwards compatibility tests.
func ValidateStepFunc(f interface{}) error {
	return validateStepFunc(f)
}

func TestValidateStepFunc_Context(t *testing.T) {
	testCases := map[string]interface{}{
		"function with invalid first argument": func(int, context.Context) {},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			if err := gobdd.ValidateStepFunc(testCase); err == nil {
				t.Errorf("the test should fail for the function")
			}
		})
	}
}

func TestValidateStepFunc_ValidFunction_Context(t *testing.T) {
	if err := gobdd.ValidateStepFunc(func(context.Context) {}); err != nil {
		t.Errorf("the test should NOT fail for the function: %s", err)
	}
}

func TestValidateStepFunc_ReturnContext_Context(t *testing.T) {
	err := gobdd.ValidateStepFunc(func(gobdd.StepTest, context.Context) context.Context { return context.Context{} })
	if err != nil {
		t.Errorf("step function returning a context should NOT fail validation: %s", err)
	}
}
