package awslib

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"k8s.io/klog/v2"
	"regexp"
)

type EKSClusterNameProvider struct {
}

const (
	asgPattern     = "^kubernetes.io/cluster/(.*)$"
	ClusterNameTag = "eks:cluster-name"
)

var (
	asgCompiledRegex = regexp.MustCompile(asgPattern)
)

type ClusterNameProvider interface {
	GetClusterName(ctx context.Context, cfg aws.Config, instanceId string) (string, error)
}

func (provider EKSClusterNameProvider) GetClusterName(ctx context.Context, cfg aws.Config, instanceId string) (string, error) {
	// With EKS, there is no data source that can guarantee to return the cluster name.
	// Therefore, we need to try multiple ways to find the cluster name.
	// First, try to find the cluster name from the instance tag
	// This is the most reliable way to find the cluster name
	// However, this method will work only on new EKS clusters
	// In that case, we will try to find the cluster name from the autoscaling group
	svc1 := ec2.NewFromConfig(cfg)
	clusterName, err := provider.getClusterNameFromInstanceTag(ctx, svc1, instanceId)
	if err != nil {
		return "", fmt.Errorf("failed to get cluster name from the instance tag: %v", err)
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

func (provider EKSClusterNameProvider) getClusterNameFromInstanceTag(ctx context.Context, svc *ec2.Client, instanceId string) (string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceId},
	}

	for {
		r, err := svc.DescribeInstances(ctx, input)
		if err != nil {
			klog.Errorf("failed to describe cluster: %v", err)
			return "", err
		}

		// ClusterName will be found in the autoscaling group tag with "owned" value.
		// Find the autoscaling group that has this instance, then get the cluster name from the tag of that group
		for _, reservation := range r.Reservations {
			for _, instance := range reservation.Instances {
				if *instance.InstanceId == instanceId {
					for _, tag := range instance.Tags {
						if *tag.Key == ClusterNameTag {
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

func (provider EKSClusterNameProvider) getClusterNameFromAutoscalingGroup(ctx context.Context, cfg aws.Config, instanceId string) (string, error) {
	klog.Infof("attempting to find cluster name from autoscaling group")
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
		for _, autoscalingGroup := range r.AutoScalingGroups {
			for _, instance := range autoscalingGroup.Instances {
				if *instance.InstanceId == instanceId {
					for _, tag := range autoscalingGroup.Tags {
						tagB := []byte(*tag.Key)
						if *tag.Value == "owned" && asgCompiledRegex.Match(tagB) {
							groups := asgCompiledRegex.FindSubmatch(tagB)
							clusterName := string(groups[1][:])
							klog.Infof("found cluster name from autoscaling group: '%v'", clusterName)
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
	return "", fmt.Errorf("cluster name not found from autoscaling groups")
}
