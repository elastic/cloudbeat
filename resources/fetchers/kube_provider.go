package fetchers

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/beats/v7/libbeat/logp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

type K8sResource struct {
	Data interface{}
}

const k8sObjMetadataField = "ObjectMeta"

func GetKubeData(watchers []kubernetes.Watcher) []PolicyResource {
	ret := make([]PolicyResource, 0)

	for _, watcher := range watchers {
		rs := watcher.Store().List()

		for _, r := range rs {
			nullifyManagedFields(r)
			resource, ok := r.(kubernetes.Resource)

			if !ok {
				logp.L().Errorf("Bad resource: %#v does not implement kubernetes.Resource", r)
				continue
			}

			err := addTypeInformationToKubeResource(resource)
			if err != nil {
				logp.L().Errorf("Bad resource: %w", err)
				continue
			} // See https://github.com/kubernetes/kubernetes/issues/3030

			ret = append(ret, K8sResource{resource})
		}
	}

	return ret
}

func (r K8sResource) GetID() string {
	k8sObj := reflect.ValueOf(r.Data)
	metadata, ok := k8sObj.FieldByName(k8sObjMetadataField).Interface().(metav1.ObjectMeta)
	if !ok {
		fmt.Errorf("failed to retrieve object metadata")
		return ""
	}

	uid := metadata.UID
	return string(uid)
}

func (r K8sResource) GetData() interface{} {
	return r.Data
}

// nullifyManagedFields ManagedFields field contains fields with dot that prevent from elasticsearch to index
// the events.
func nullifyManagedFields(resource interface{}) {
	switch val := resource.(type) {
	case *kubernetes.Pod:
		val.ManagedFields = nil
	case *kubernetes.Secret:
		val.ManagedFields = nil
	case *kubernetes.Role:
		val.ManagedFields = nil
	case *kubernetes.RoleBinding:
		val.ManagedFields = nil
	case *kubernetes.ClusterRole:
		val.ManagedFields = nil
	case *kubernetes.ClusterRoleBinding:
		val.ManagedFields = nil
	case *kubernetes.PodSecurityPolicy:
		val.ManagedFields = nil
	case *kubernetes.NetworkPolicy:
		val.ManagedFields = nil
	}
}
