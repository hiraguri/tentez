package tentez

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestAwsListenerRule_execSwitch(t *testing.T) {
	cases := []struct {
		isError         bool
		awsListenerRule AwsListenerRule
		weight          Weight
		elbv2Mock       elbv2Mock
	}{
		{
			isError: false,
			awsListenerRule: AwsListenerRule{
				Name:   "success",
				Target: "validTarget",
				Switch: Switch{
					Old: "oldTarget",
					New: "newTarget",
				},
			},
		},
		{
			isError: true,
			awsListenerRule: AwsListenerRule{
				Name:   "success",
				Target: "validTarget",
				Switch: Switch{
					Old: "oldTarget",
					New: "newTarget",
				},
			},
			weight: Weight{
				Old: 30,
				New: 70,
			},
			elbv2Mock: elbv2Mock{
				ModifyRuleError: fmt.Errorf("error"),
			},
		},
	}

	for _, c := range cases {
		err := c.awsListenerRule.execSwitch(c.weight, Config{
			client: Client{
				elbv2: c.elbv2Mock,
			},
			io: IOStreams{
				in:  bytes.NewBufferString(""),
				out: bytes.NewBufferString(""),
				err: bytes.NewBufferString(""),
			},
		})

		if c.isError != (err != nil) {
			t.Errorf("expect isError == %t, but got %v", c.isError, err)
		}
	}
}

func TestAwsListenerRules_fetchData(t *testing.T) {
	cases := []struct {
		isError          bool
		expect           interface{}
		awsListenerRules AwsListenerRules
		elbv2Mock        elbv2Mock
	}{
		{
			isError: false,
			expect: struct {
				AwsListenerRules []AwsListenerRuleData `yaml:"aws_listener_rules"`
			}{},
			awsListenerRules: AwsListenerRules{
				AwsListenerRule{
					Name:   "success",
					Target: "validTarget",
					Switch: Switch{
						Old: "oldTarget",
						New: "newTarget",
					},
				},
			},
		},
		{
			isError: true,
			expect:  nil,
			awsListenerRules: AwsListenerRules{
				AwsListenerRule{
					Name:   "success",
					Target: "validTarget",
					Switch: Switch{
						Old: "oldTarget",
						New: "newTarget",
					},
				},
			},
			elbv2Mock: elbv2Mock{
				DescribeRulesError: fmt.Errorf("error"),
			},
		},
	}

	for _, c := range cases {
		got, gotErr := c.awsListenerRules.fetchData(Config{
			client: Client{
				elbv2: c.elbv2Mock,
			},
			io: IOStreams{
				in:  bytes.NewBufferString(""),
				out: bytes.NewBufferString(""),
				err: bytes.NewBufferString(""),
			},
		})

		expectedYaml, err := yaml.Marshal(c.expect)
		if err != nil {
			t.Errorf("cannot yaml.Marshal: %s", err)
		}
		gotYaml, err := yaml.Marshal(got)
		if err != nil {
			t.Errorf("cannot yaml.Marshal: %s", err)
		}

		if !reflect.DeepEqual(expectedYaml, gotYaml) {
			t.Errorf("expect %+v, but got %+v", c.expect, got)
		}
		if c.isError != (gotErr != nil) {
			t.Errorf("expect isError == %t, but got %v", c.isError, gotErr)
		}
	}
}
