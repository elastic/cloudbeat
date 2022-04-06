package aws_providers

type MockAwsCredentialsGetter func() AwsFetcherConfig

func (m MockAwsCredentialsGetter) GetAwsCredentials() AwsFetcherConfig {
	return m.GetAwsCredentials()
}
