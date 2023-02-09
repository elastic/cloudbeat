package configservice

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	configSDK "github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

const awsAccountId = "dummy-account-id"

func TestProvider_DescribeConfigRecorders(t *testing.T) {
	tests := []struct {
		name            string
		mockClient      func() Client
		regions         []string
		wantErr         bool
		expectedResults int
	}{
		{
			name: "Should return a config without recorders",
			mockClient: func() Client {
				m := MockClient{}
				m.On("DescribeConfigurationRecorders", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecordersOutput{}, nil)
				return &m
			},
			regions:         []string{"us-east-1"},
			wantErr:         false,
			expectedResults: 1,
		},
		{
			name: "Should not return a config due to error",
			mockClient: func() Client {
				m := MockClient{}
				m.On("DescribeConfigurationRecorders", mock.Anything, mock.Anything).Return(nil, errors.New("API_ERROR"))
				return &m
			},
			regions:         []string{"us-east-1"},
			wantErr:         true,
			expectedResults: 0,
		},
		{
			name: "Should return config resources",
			mockClient: func() Client {
				m := MockClient{}
				m.On("DescribeConfigurationRecorders", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecordersOutput{
					ConfigurationRecorders: []types.ConfigurationRecorder{{Name: aws.String("test1")}}}, nil).Once()
				m.On("DescribeConfigurationRecorders", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecordersOutput{
					ConfigurationRecorders: []types.ConfigurationRecorder{{Name: aws.String("test2")}}}, nil).Once()

				m.On("DescribeConfigurationRecorderStatus", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecorderStatusOutput{
					ConfigurationRecordersStatus: []types.ConfigurationRecorderStatus{{Name: aws.String("test1")}}}, nil).Once()
				m.On("DescribeConfigurationRecorderStatus", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecorderStatusOutput{
					ConfigurationRecordersStatus: []types.ConfigurationRecorderStatus{{Name: aws.String("test2")}}}, nil).Once()

				return &m
			},
			regions:         []string{"us-east-1", "us-east-2"},
			wantErr:         false,
			expectedResults: 2,
		},
		{
			name: "Should return config resources from a single region",
			mockClient: func() Client {
				m := MockClient{}
				m.On("DescribeConfigurationRecorders", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecordersOutput{
					ConfigurationRecorders: []types.ConfigurationRecorder{{Name: aws.String("test1")}}}, nil).Once()
				m.On("DescribeConfigurationRecorders", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecordersOutput{
					ConfigurationRecorders: []types.ConfigurationRecorder{{Name: aws.String("test2")}}}, nil).Once()

				m.On("DescribeConfigurationRecorderStatus", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecorderStatusOutput{
					ConfigurationRecordersStatus: []types.ConfigurationRecorderStatus{{Name: aws.String("test1")}}}, nil).Once()
				m.On("DescribeConfigurationRecorderStatus", mock.Anything, mock.Anything).Return(nil, errors.New("API_ERROR")).Once()

				return &m
			},
			regions:         []string{"us-east-1", "us-east-2"},
			wantErr:         true,
			expectedResults: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				log:          logp.NewLogger("configservice_provider_test"),
				awsAccountId: awsAccountId,
				clients:      testhelper.CreateMockClients[Client](tt.mockClient(), tt.regions),
			}

			got, err := p.DescribeConfigRecorders(context.TODO())
			if (err != nil) != tt.wantErr {
				t.Errorf("DescribeConfigRecorders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.expectedResults, len(got))
		})
	}
}
