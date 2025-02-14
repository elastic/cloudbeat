// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package awslib

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"k8s.io/klog/v2"
)

type EKSClusterNameProvider struct{}

const (
	asgPattern                               = "^kubernetes.io/cluster/(.*)$"
	clusterNameTag                           = "eks:cluster-name"
	numberOfAutoScalingGroups                = 100
	numberOfIterationsInEachAutoScalingGroup = 100
)

var (
	asgCompiledRegex = regexp.MustCompile(asgPattern)
)

type EKSClusterNameProviderAPI interface {
	GetClusterName(ctx context.Context, cfg aws.Config, instanceId string) (string, error)
}

func (provider EKSClusterNameProvider) GetClusterName(ctx context.Context, cfg aws.Config, instanceId string) (string, error) {
	// With EKS, there is no data source that can guarantee to return the cluster name.
	// Therefore, we need to try multiple ways to find the cluster name.
	// First, try to extract the cluster name from the instance tags.
	// This is the most reliable way to find the cluster name.
	// However, this method will work only on new EKS clusters.
	// Therefore, if the tag was not found we will try to extract the cluster name from the autoscaling group.
	clusterName, err := provider.getClusterNameFromInstanceTags(ctx, cfg, instanceId)
	if err != nil {
		return "", fmt.Errorf("failed to get cluster name from the instance tags: %v", err)
	}
	if clusterName != "" {
		return clusterName, nil
	}
	clusterName, err = provider.getClusterNameFromAutoscalingGroup(ctx, cfg, instanceId)
	if err != nil {
		return "", fmt.Errorf("failed to get cluster name from the Auto-scaling group: %v", err)
	}

	return clusterName, nil
}

func (provider EKSClusterNameProvider) getClusterNameFromInstanceTags(ctx context.Context, cfg aws.Config, instanceId string) (string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceId},
	}
	svc := ec2.NewFromConfig(cfg)

	for {
		r, err := svc.DescribeInstances(ctx, input)
		if err != nil {
			return "", fmt.Errorf("failed to describe instance required for cluster name detection: %v", err)
		}

		// Look for the cluster name tag in the instance tags
		for _, reservation := range r.Reservations {
			for _, instance := range reservation.Instances {
				if *instance.InstanceId == instanceId {
					for _, tag := range instance.Tags {
						if strings.HasSuffix(*tag.Key, clusterNameTag) {
							return *tag.Value, nil
						}
					}
				}
			}
		}

		if r.NextToken == nil {
			break
		}
		input.NextToken = r.NextToken
	}
	return "", nil
}

//revive:disable-next-line:cognitive-complexity
func (provider EKSClusterNameProvider) getClusterNameFromAutoscalingGroup(ctx context.Context, cfg aws.Config, instanceId string) (string, error) {
	svc := autoscaling.NewFromConfig(cfg)
	input := &autoscaling.DescribeAutoScalingGroupsInput{}

	for {
		r, err := svc.DescribeAutoScalingGroups(ctx, input)
		if err != nil {
			klog.Errorf("ec2 describe-autoscaling-group: %v", err)
			return "", err
		}

		// ClusterName will be found in the autoscaling group tag with "owned" value.
		// Find the autoscaling group that has this instance, then get the cluster name from the tag of that group
		// We wish to limit the search and preform a best effort search, therefore we will limit the number of iterations
		for scalingGroupNumber, autoscalingGroup := range r.AutoScalingGroups {
			if scalingGroupNumber > numberOfAutoScalingGroups {
				break
			}
			for numberOfInstance, instance := range autoscalingGroup.Instances {
				if numberOfInstance > numberOfIterationsInEachAutoScalingGroup {
					break
				}
				if *instance.InstanceId == instanceId {
					for _, tag := range autoscalingGroup.Tags {
						stringifyTag := *tag.Key
						if *tag.Value == "owned" && asgCompiledRegex.MatchString(stringifyTag) {
							groups := asgCompiledRegex.FindStringSubmatch(stringifyTag)
							clusterName := groups[1]
							return clusterName, nil
						}
					}
				}
			}
		}

		if r.NextToken == nil {
			break
		}
		input.NextToken = r.NextToken
	}
	return "", errors.New("cluster name not found from autoscaling groups")
}
