package tentez

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2Types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

type AwsListener struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
	Switch Switch `yaml:"switch"`
}

type AwsListeners []AwsListener

type AwsListenerData struct {
	ListnerArn string                `yaml:"target"`
	Weights    []AwsTargetGroupTuple `yaml:"weights"`
}

func (l AwsListener) execSwitch(weight Weight, client Client) error {
	// avoid rate limit
	time.Sleep(1 * time.Second)

	_, err := client.elbv2.ModifyListener(context.TODO(), &elbv2.ModifyListenerInput{
		ListenerArn: aws.String(l.Target),
		DefaultActions: []elbv2Types.Action{
			{
				Type: "forward",
				ForwardConfig: &elbv2Types.ForwardActionConfig{
					TargetGroups: []elbv2Types.TargetGroupTuple{
						{
							TargetGroupArn: aws.String(l.Switch.Old),
							Weight:         aws.Int32(weight.Old),
						},
						{
							TargetGroupArn: aws.String(l.Switch.New),
							Weight:         aws.Int32(weight.New),
						},
					},
				},
			},
		},
	})

	return err
}

func (l AwsListener) getName() string {
	return l.Name
}

func (ls AwsListeners) fetchData(client Client) (interface{}, error) {
	if len(ls) == 0 {
		return nil, nil
	}

	listenerArns := []string{}
	for _, listener := range ls {
		listenerArns = append(listenerArns, listener.Target)
	}

	listenersData, err := client.elbv2.DescribeListeners(context.TODO(), &elbv2.DescribeListenersInput{
		ListenerArns: listenerArns,
	})
	if err != nil {
		return nil, err
	}

	res := struct {
		AwsListeners []AwsListenerData `yaml:"aws_listeners"`
	}{
		AwsListeners: []AwsListenerData{},
	}

	for _, listenerData := range listenersData.Listeners {
		for _, action := range listenerData.DefaultActions {
			targetGroupTuples := []AwsTargetGroupTuple{}
			for _, tgTuple := range action.ForwardConfig.TargetGroups {
				targetGroupTuples = append(targetGroupTuples, AwsTargetGroupTuple{
					TargetGroupArn: aws.ToString(tgTuple.TargetGroupArn),
					Weight:         aws.ToInt32(tgTuple.Weight),
				})
			}

			res.AwsListeners = append(res.AwsListeners, AwsListenerData{
				ListnerArn: *listenerData.ListenerArn,
				Weights:    targetGroupTuples,
			})
		}
	}

	return res, nil
}

func (ls AwsListeners) targetsSlice() (targets []Target) {
	for _, target := range ls {
		targets = append(targets, target)
	}
	return targets
}
