package waf

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FindSubscribedRuleGroupByNameOrMetricName(conn *waf.WAF, name string, metricName string) (*waf.SubscribedRuleGroupSummary, error) {
	if name == "" && metricName == "" {
		return nil, errors.New("must specify either name or metricName")
	}

	hasName := name != ""
	hasMetricName := metricName != ""
	hasMatch := false

	input := &waf.ListSubscribedRuleGroupsInput{}

	matchingRuleGroup := &waf.SubscribedRuleGroupSummary{}

	for {
		output, err := conn.ListSubscribedRuleGroups(input)

		if tfawserr.ErrCodeContains(err, waf.ErrCodeNonexistentItemException) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, ruleGroup := range output.RuleGroups {
			respName := aws.StringValue(ruleGroup.Name)
			respMetricName := aws.StringValue(ruleGroup.MetricName)

			if hasName && respName != name {
				continue
			}
			if hasMetricName && respMetricName != metricName {
				continue
			}
			if hasName && hasMetricName && (name != respName || metricName != metricName) {
				continue
			}
			// Previous conditionals catch all non-matches
			if hasMatch {
				return nil, fmt.Errorf("multiple matches found for name %s and metricName %s", name, metricName)
			}

			matchingRuleGroup = ruleGroup
			hasMatch = true
		}

		if output.NextMarker == nil {
			break
		}
		input.NextMarker = output.NextMarker
	}

	if !hasMatch {
		return nil, fmt.Errorf("no matches found for name %s and metricName %s", name, metricName)
	}

	return matchingRuleGroup, nil
}
