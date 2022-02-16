package fetchers

import (
	"context"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/beats/v7/libbeat/logp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"regexp"
	"sync"

	"github.com/elastic/beats/v7/cloudbeat/resources"
)

const ECRType = "aws-ecr"
const PrivateRepoRegexTemplate = "^%s\\.dkr\\.ecr\\.%s\\.amazonaws\\.com\\/([-\\w]+)[:,@]?"
const PublicRepoRegex = "public\\.ecr\\.aws\\/\\w+\\/([\\w-]+)\\:?"

type ECRFetcher struct {
	cfg               ECRFetcherConfig
	ecrProvider       *ECRProvider
	kubeClient        k8s.Interface
	kubeInitOnce      sync.Once
	repoRegexMatchers []*regexp.Regexp
}

type ECRFetcherConfig struct {
	resources.BaseFetcherConfig
	Kubeconfig string `config:"Kubeconfig"`
}

func NewECRFetcher(awsConfig resources.AwsFetcherConfig, cfg ECRFetcherConfig) (resources.Fetcher, error) {
	ecr := NewEcrProvider(awsConfig.Config)

	privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, *awsConfig.AccountID, awsConfig.Config.Region)

	return &ECRFetcher{
		cfg:         cfg,
		ecrProvider: ecr,
		repoRegexMatchers: []*regexp.Regexp{
			regexp.MustCompile(privateRepoRegex),
			regexp.MustCompile(PublicRepoRegex),
		},
	}, nil
}

// Fetch This function should be called once per cluster (Leader Election)
func (f ECRFetcher) Fetch(ctx context.Context) ([]resources.FetcherResult, error) {
	var err error
	f.kubeInitOnce.Do(func() {
		f.kubeClient, err = kubernetes.GetKubernetesClient(f.cfg.Kubeconfig, kubernetes.KubeClientOptions{})
	})
	if err != nil {
		// Reset watcherlock if the watchers could not be initiated.
		watcherlock = sync.Once{}
		return nil, fmt.Errorf("could not initate Kubernetes watchers: %w", err)
	}

	return f.getData(ctx)
}

func (f ECRFetcher) getData(ctx context.Context) ([]resources.FetcherResult, error) {
	podsAwsRepositories, err := f.getAwsPodRepositories(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]resources.FetcherResult, 0)

	ecrRepositories, err := f.ecrProvider.DescribeRepositories(ctx, podsAwsRepositories)
	results = append(results, resources.FetcherResult{
		Type:     ECRType,
		Resource: ecrRepositories,
	})

	return results, err
}

func (f ECRFetcher) getAwsPodRepositories(ctx context.Context) ([]string, error) {
	podsList, err := f.kubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		logp.Err("Failed to get pods  - %+v", err)
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

func (f ECRFetcher) Stop() {
}
