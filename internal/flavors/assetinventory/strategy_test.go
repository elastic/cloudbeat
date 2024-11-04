package assetinventory

import (
	context "context"
	"testing"
	"time"

	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
)

func TestStrategyPicks(t *testing.T) {
	testCases := []struct {
		name        string
		cfg         *config.Config
		expectedErr string
	}{
		{
			"expected error: asset_inventory_provider not set",
			&config.Config{},
			"missing config.v1.asset_inventory_provider",
		},
		{
			"expected error: unsupported provider",
			&config.Config{
				AssetInventoryProvider: "NOPE",
			},
			"unsupported Asset Inventory provider \"NOPE\"",
		},
		{
			"expected success: Azure",
			&config.Config{
				AssetInventoryProvider: config.ProviderAzure,
			},
			"",
		},
		{
			"expected error: GCP missing account type",
			&config.Config{
				AssetInventoryProvider: config.ProviderGCP,
			},
			"invalid gcp account type",
		},
		{
			"expected success: GCP",
			&config.Config{
				AssetInventoryProvider: config.ProviderGCP,
				CloudConfig: config.CloudConfig{
					Gcp: config.GcpConfig{
						AccountType: config.SingleAccount,
					},
				},
			},
			"",
		},
		{
			"expected error: AWS unsupported account type",
			&config.Config{
				AssetInventoryProvider: config.ProviderAWS,
				CloudConfig: config.CloudConfig{
					Aws: config.AwsConfig{
						AccountType: "NOPE",
					},
				},
			},
			"unsupported account_type: \"NOPE\"",
		},
		{
			"expected success: AWS",
			&config.Config{
				AssetInventoryProvider: config.ProviderAWS,
				CloudConfig: config.CloudConfig{
					Aws: config.AwsConfig{
						AccountType: config.SingleAccount,
						Cred: aws.ConfigAWS{
							AccessKeyID:     "key",
							SecretAccessKey: "key",
						},
					},
				},
			},
			"STS: GetCallerIdentity",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := strategy{
				logger: logp.NewLogger("strategy_test"),
				cfg:    tc.cfg,
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			obj, err := s.NewAssetInventory(ctx, nil)
			if tc.expectedErr != "" {
				assert.Equal(t, inventory.AssetInventory{}, obj)
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
