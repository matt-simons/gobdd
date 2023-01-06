package gobdd

import (
	"context"
	"errors"
	"reflect"
)

func validateStepFunc(f interface{}) error {
	value := reflect.ValueOf(f)
	if value.Kind() != reflect.Func {
		return errors.New("the parameter should be a function")
	}

	if value.Type().NumIn() < 1 {
		return errors.New("the function should have Context as the first argument")
	}

	val := value.Type().In(0)

	testingInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !val.Implements(testingInterface) {
		return errors.New("the function should have Context as the first argument")
	}

	return nil
}
