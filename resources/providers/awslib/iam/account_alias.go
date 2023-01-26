package iam

import (
	"context"
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
)

func (p Provider) GetAccountAlias(ctx context.Context) (string, error) {
	aliases, err := p.client.ListAccountAliases(ctx, &iamsdk.ListAccountAliasesInput{})
	if err != nil {
		return "", err
	}

	if len(aliases.AccountAliases) > 0 {
		return aliases.AccountAliases[0], nil
	}

	return "", nil
}
