package transformer

import (
	"github.com/gofrs/uuid"
)

type ResourceTypeMetadata struct {
	CycleMetadata
	Type string
}

type ResourceMetadata struct {
	ResourceTypeMetadata
	ResourceId string
}

type CycleMetadata struct {
	CycleId uuid.UUID
}
