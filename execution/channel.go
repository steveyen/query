//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package execution

import (
	"github.com/couchbaselabs/query/value"
)

// Dummy operator that simply wraps an item channel.
type Channel struct {
	base
}

func NewChannel() *Channel {
	rv := &Channel{
		base: newBase(),
	}

	rv.output = rv
	return rv
}

func (this *Channel) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitChannel(this)
}

func (this *Channel) Copy() Operator {
	return &Channel{
		this.base.copy(),
	}
}

// This operator is a no-op. It simply provides a shared itemChannel.
func (this *Channel) RunOnce(context *Context, parent value.Value) {
	this.once.Do(func() {
		defer context.Recover()       // Recover from any panic
		defer close(this.itemChannel) // Broadcast that I have stopped
		defer this.notify()           // Notify that I have stopped

		// Block until stopped
		<-this.StopChannel()
	})
}
