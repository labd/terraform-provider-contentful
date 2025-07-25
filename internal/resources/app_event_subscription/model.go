package app_event_subscription

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
)

// AppEventSubscription is the main resource schema data
type AppEventSubscription struct {
	ID              types.String   `tfsdk:"id"`
	AppDefinitionID types.String   `tfsdk:"app_definition_id"`
	TargetUrl       types.String   `tfsdk:"target_url"`
	Topics          []types.String `tfsdk:"topics"`
}

func (a *AppEventSubscription) Draft() *sdk.AppEventSubscriptionDraft {
	var topics []string
	for _, topic := range a.Topics {
		topics = append(topics, topic.ValueString())
	}

	app := &sdk.AppEventSubscriptionDraft{
		TargetUrl: a.TargetUrl.ValueString(),
		Topics:    topics,
	}

	return app
}

func (a *AppEventSubscription) Equal(n *AppEventSubscription) bool {
	if !a.TargetUrl.Equal(n.TargetUrl) {
		return false
	}

	if len(a.Topics) != len(n.Topics) {
		return false
	}

	for i, topic := range a.Topics {
		if !topic.Equal(n.Topics[i]) {
			return false
		}
	}

	return true
}

func Import(a *AppEventSubscription, n *sdk.AppEventSubscription) {
	a.TargetUrl = types.StringValue(n.TargetUrl)

	var topics []types.String
	for _, topic := range n.Topics {
		topics = append(topics, types.StringValue(topic))
	}

	a.Topics = topics
}
