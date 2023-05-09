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

// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package dev

import (
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/ec2"
)

type SecurityGroupDevMod struct{}

func (m *SecurityGroupDevMod) Modify(template *cloudformation.Template) error {
	securityGroups, err := template.GetEC2SecurityGroupWithName("ElasticAgentSecurityGroup")
	if err != nil {
		return err
	}
	securityGroups.GroupDescription = "Allow SSH from anywhere"
	securityGroups.SecurityGroupIngress = []ec2.SecurityGroup_Ingress{
		{
			IpProtocol: "tcp",
			FromPort:   cloudformation.Int(22),
			ToPort:     cloudformation.Int(22),
			CidrIp:     cloudformation.String("0.0.0.0/0"),
		},
	}
	return nil
}

type Ec2KeyDevMod struct{}

func (m Ec2KeyDevMod) Modify(template *cloudformation.Template) error {
	template.Parameters["KeyName"] = cloudformation.Parameter{
		Type:        "AWS::EC2::KeyPair::KeyName",
		Description: cloudformation.String("SSH Keypair to login to the instance"),
	}

	ec2Instance, err := template.GetEC2InstanceWithName("ElasticAgentEc2Instance")
	if err != nil {
		return err
	}

	ec2Instance.KeyName = cloudformation.RefPtr("KeyName")
	return nil
}
