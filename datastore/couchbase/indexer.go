//  Copyright (c) 2016 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the
//  License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an "AS
//  IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language
//  governing permissions and limitations under the License.

package couchbase

import (
	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/errors"
)

// A registry of couchbase indexer providers, keyed by indexer
// provider name, like "fts".
var indexerProviderRegistry map[string]IndexerProvider

// IndexerProvider represents the methods that a couchbase indexer
// implementation must implement.
type IndexerProvider struct {
	Create func(clusterUrl, namespace, name string) (datastore.Indexer, errors.Error)
}

// RegisterIndexerProvider allows external integrations to register
// additional couchbase indexer implementations.  Registration should
// only happen during process init()'ialization.
func RegisterIndexerProvider(name string, ip IndexerProvider) error {
	if indexerProviderRegistry == nil {
		indexerProviderRegistry = map[string]IndexerProvider{}
	}
	indexerProviderRegistry[name] = ip
	return nil
}
