// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package resources

import (
	"context"
	"encoding/gob"
	"fmt"
	"reflect"
	"testing"
	"time"

	"go.uber.org/goleak"
)

const (
	duration     = 10 * time.Second
	fetcherCount = 10
)

func TestDataRun(t *testing.T) {
	gob.Register(NumberResource{})
	opts := goleak.IgnoreCurrent()

	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	defer goleak.VerifyNone(t, opts)

	reg := NewFetcherRegistry()
	registerNFetchers(t, reg, fetcherCount)
	d, err := NewData(duration, reg)
	if err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = d.Run(ctx)
	if err != nil {
		return
	}
	defer d.Stop(ctx, cancel)

	o := d.Output()
	state := <-o

	if len(state) < fetcherCount {
		t.Errorf("expected %d keys but got %d", fetcherCount, len(state))
	}

	for i := 0; i < fetcherCount; i++ {
		key := fmt.Sprint(i)

		val, ok := state[key]
		if !ok {
			t.Errorf("expected key %s but not found", key)
		}

		if !reflect.DeepEqual(val, fetchValue(i)) {
			t.Errorf("expected key %s to have value %v but got %v", key, fetchValue(i), val)
		}
	}
}
