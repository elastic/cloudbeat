package inventory_test

import (
	"context"
	awsbeat "github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/internal/inventory"
	awsinventory "github.com/elastic/cloudbeat/internal/inventory/aws"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
	"testing"
	"time"
)

func TestInventory(t *testing.T) {
	ctx := context.Background()

	err := logp.DevelopmentSetup()
	if err != nil {
		return
	}
	logger := logp.NewLogger("asset_inventory")

	awsConfig, err := awsbeat.InitializeAWSConfig(awsbeat.ConfigAWS{})

	if err != nil {
		t.Errorf("Error initializing aws config %v", err)
		return
	}

	identityProvider := awslib.IdentityProvider{}
	awsIdentity, err := identityProvider.GetIdentity(ctx, awsConfig)
	if err != nil {
		t.Errorf("Error initializing aws identity %v", err)
		return
	}

	fetchers := awsinventory.AwsFetchers(logger, awsIdentity, awsConfig)
	inv := inventory.NewAssetInventory(logger, fetchers)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	inv.BuildInventory(ctx)
}
