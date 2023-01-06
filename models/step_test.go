package models

import (
	"context"

	messages "github.com/cucumber/messages/go/v21"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Running Steps", func() {

	Context("Running Basic Steps", Ordered, func() {
		scheme := &Scheme{}

		BeforeAll(func() {
			Expect(scheme.Register(basicGoodStep)).Should(Succeed())
		})

		It("should run a basic good step", func() {
			stepDoc := &messages.Step{
				Text: "a word",
			}
			step, err := NewStep(stepDoc, scheme)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(step).ShouldNot(BeNil())

			step.Run(context.TODO())
			Expect(step.Execution.Result).Should(Equal(Passed))
			Expect(step.Execution.Err).ShouldNot(HaveOccurred())
			Expect(step.Execution.StartTime).Should(BeTemporally("<", step.Execution.EndTime))
		})
	})

})
