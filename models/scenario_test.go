package models

import (
	"context"

	messages "github.com/cucumber/messages/go/v21"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Running Scenarios", func() {

	Context("Running Basic Scenarios", Ordered, func() {
		scheme := &Scheme{}

		BeforeAll(func() {
			Expect(scheme.Register(basicGoodStep)).Should(Succeed())
		})

		It("should run a basic good step", func() {

			backgroundDoc := &messages.Background{
				Steps: []*messages.Step{
					{
						Text: "a word",
					},
				},
			}
			scenarioDoc := &messages.Scenario{
				Steps: []*messages.Step{
					{
						Text: "a word",
					},
				},
			}
			scenario, err := NewScenario(backgroundDoc, scenarioDoc, scheme)
			Expect(err).ShouldNot(HaveOccurred())

			scenario.Run(context.TODO())
			Expect(scenario.Steps[0].Execution.Result).Should(Equal(Passed))
			Expect(scenario.Steps[1].Execution.Result).Should(Equal(Passed))
		})
	})

})
