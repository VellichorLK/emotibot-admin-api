package data

import (
	"errors"
)

var (
	ErrESNotAcknowledged      = errors.New("Elasticsearch: Not acknowledged")
	ErrESIndexNotExisted      = errors.New("Elasticsearch: Index not exists")
	ErrESTermsNotFound        = errors.New("Elasticsearch: Terms not found")
	ErrESMaxBucketNotFound    = errors.New("Elasticsearch: Max bucket not found")
	ErrESMinBucketNotFound    = errors.New("Elasticsearch: Min bucket not found")
	ErrESCardinalityNotFound  = errors.New("Elasticsearch: Cardinality not found")
	ErrESTopHitNotFound       = errors.New("Elasticsearch: TopHit not found")
	ErrESNameBucketNotFound   = errors.New("Elasticsearch: Name bucket not found")
	ErrNotInit                = errors.New("Elasticsearch client is not inited successfully")
)
