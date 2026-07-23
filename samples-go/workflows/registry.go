package workflows

import (
	"github.com/superdurable/iwf-golang-samples/workflows/engagement"
	"github.com/superdurable/iwf-golang-samples/workflows/microservices"
	"github.com/superdurable/iwf-golang-samples/workflows/moneytransfer"
	"github.com/superdurable/iwf-golang-samples/workflows/polling"
	"github.com/superdurable/iwf-golang-samples/workflows/service"
	"github.com/superdurable/iwf-golang-samples/workflows/subscription"
	"github.com/superdurable/iwf/sdk-go/iwf"
)

var registry = iwf.NewRegistry()

func init() {

	svc := service.NewMyService()

	err := registry.AddWorkflows(
		subscription.NewSubscriptionWorkflow(svc),
		engagement.NewEngagementWorkflow(svc),
		microservices.NewMicroserviceOrchestrationWorkflow(svc),
		moneytransfer.NewMoneyTransferWorkflow(svc),
		polling.NewPollingWorkflow(svc),
	)
	if err != nil {
		panic(err)
	}
}

func GetRegistry() iwf.Registry {
	return registry
}
