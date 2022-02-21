package fetchers

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/beats/v7/libbeat/logp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"regexp"
)

const ECRType = "aws-ecr"
const PrivateRepoRegexTemplate = "^%s\\.dkr\\.ecr\\.%s\\.amazonaws\\.com\\/([-\\w]+)[:,@]?"
const PublicRepoRegex = "public\\.ecr\\.aws\\/\\w+\\/([\\w-]+)\\:?"

type ECRFetcher struct {
	cfg               ECRFetcherConfig
	ecrProvider       *ECRProvider
	kubeClient        k8s.Interface
	repoRegexMatchers []*regexp.Regexp
}

type ECRFetcherConfig struct {
	BaseFetcherConfig
	Kubeconfig string `config:"Kubeconfig"`
}

type EcrRepositories []ecr.Repository

type ECRResource struct {
	EcrRepositories
}

func NewECRFetcher(awsCfg AwsFetcherConfig, cfg ECRFetcherConfig) (Fetcher, error) {
	ecrProvider := NewEcrProvider(awsCfg.Config)

	privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, *awsCfg.AccountID, awsCfg.Config.Region)
	kubeClient, err := kubernetes.GetKubernetesClient(cfg.Kubeconfig, kubernetes.KubeClientOptions{})

	if err != nil {
		return nil, fmt.Errorf("could not initate Kubernetes client: %w", err)
	}
	return &ECRFetcher{
		cfg:         cfg,
		ecrProvider: ecrProvider,
		kubeClient:  kubeClient,
		repoRegexMatchers: []*regexp.Regexp{
			regexp.MustCompile(privateRepoRegex),
			regexp.MustCompile(PublicRepoRegex),
		},
	}, nil
}

func (f *ECRFetcher) Stop() {
}

func (f *ECRFetcher) Fetch(ctx context.Context) ([]FetchedResource, error) {
	results := make([]FetchedResource, 0)
	podsAwsRepositories, err := f.getAwsPodRepositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve pod's aws repositories: %w", err)
	}
	ecrRepositories, err := f.ecrProvider.DescribeRepositories(ctx, podsAwsRepositories)
	if err != nil {
		return nil, fmt.Errorf("could retrieve ECR repositories: %w", err)
	}

	results = append(results, ECRResource{ecrRepositories})
	return results, err
}

func (f *ECRFetcher) getAwsPodRepositories(ctx context.Context) ([]string, error) {
	podsList, err := f.kubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("failed to get pods  - %w", err))
		return nil, err
	}

	repositories := make([]string, 0)
	for _, pod := range podsList.Items {
		for _, container := range pod.Spec.Containers {
			image := container.Image
			// Takes only aws images
			for _, matcher := range f.repoRegexMatchers {
				if matcher.MatchString(image) {
					// Extract the repository name out of the image name
					repository := matcher.FindStringSubmatch(image)[1]
					repositories = append(repositories, repository)
				}
			}
		}
	}
	return repositories, nil
}

// GetID TODO: Add resource id logic to all AWS resources
func (res ECRResource) GetID() string {
	return ""
}
func (res ECRResource) GetData() interface{} {
	return res
}
