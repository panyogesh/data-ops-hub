package iplocate

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func GetAddressFromIP(ctx workflow.Context, name string) (string, error) {

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second, //amount of time that must elapse before the first retry occurs
			MaximumInterval:    time.Minute, //maximum interval between retries
			BackoffCoefficient: 2,           //how much the retry interval increases
			// MaximumAttempts: 5, // Uncomment this if you want to limit attempts
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	var iPActivties IPActivties
	var ip string

	err := workflow.ExecuteActivity(ctx, iPActivties.GetIP).Get(ctx, &ip)
	if err != nil {
		return "", fmt.Errorf("Failed to get IP: %s", err)
	}

	var locationInfo string
	err = workflow.ExecuteLocalActivity(ctx, ipActivties.GetLocationInfo, ip).Get(ctx, &locationInfo))

	return "", nil
}
