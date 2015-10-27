//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package algebra

import (
	"encoding/json"

	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
)

/*
Nesting is conceptually the inverse of unnesting. Nesting performs a
join across two keyspaces (or a keyspace with itself). But instead of
producing a cross-product of the left and right hand inputs, a single
result is produced for each left hand input, while the corresponding
right hand inputs are collected into an array and nested as a single
array-valued field in the result object.
*/
type Nest struct {
	left  FromTerm
	right *KeyspaceTerm
	outer bool
}

func NewNest(left FromTerm, outer bool, right *KeyspaceTerm) *Nest {
	return &Nest{left, right, outer}
}

func (this *Nest) Accept(visitor NodeVisitor) (interface{}, error) {
	return visitor.VisitNest(this)
}

/*
Maps the right input of the nest if the left is mapped
successfully.
*/
func (this *Nest) MapExpressions(mapper expression.Mapper) (err error) {
	err = this.left.MapExpressions(mapper)
	if err != nil {
		return
	}

	return this.right.MapExpressions(mapper)
}

/*
   Returns all contained Expressions.
*/
func (this *Nest) Expressions() expression.Expressions {
	return append(this.left.Expressions(), this.right.Expressions()...)
}

/*
Returns all required privileges.
*/
func (this *Nest) Privileges() (datastore.Privileges, errors.Error) {
	privs, err := this.left.Privileges()
	if err != nil {
		return nil, err
	}

	rprivs, err := this.right.Privileges()
	if err != nil {
		return nil, err
	}

	privs.Add(rprivs)
	return privs, nil
}

/*
   Representation as a N1QL string.
*/
func (this *Nest) String() string {
	s := this.left.String()

	if this.outer {
		s += " left outer nest "
	} else {
		s += " nest "
	}

	s += this.right.toString(true)
	return s
}

/*
Qualify all identifiers for the parent expression. Checks is
a nest alias exists and if it is a duplicate alias.
*/
func (this *Nest) Formalize(parent *expression.Formalizer) (f *expression.Formalizer, err error) {
	f, err = this.left.Formalize(parent)
	if err != nil {
		return
	}

	f.Keyspace = ""
	this.right.keys, err = f.Map(this.right.keys)
	if err != nil {
		return
	}

	alias := this.Alias()
	if alias == "" {
		err = errors.NewNoTermNameError("NEST", "plan.nest.requires_name_or_alias")
		return nil, err
	}

	_, ok := f.Allowed.Field(alias)
	if ok {
		err = errors.NewDuplicateAliasError("NEST", alias, "plan.nest.duplicate_alias")
		return nil, err
	}

	f.Allowed.SetField(alias, alias)
	return
}

/*
Return the primary term in the left term of the nest clause.
*/
func (this *Nest) PrimaryTerm() FromTerm {
	return this.left.PrimaryTerm()
}

/*
Returns the nest alias of the right source.
*/
func (this *Nest) Alias() string {
	return this.right.Alias()
}

/*
Returns the left term in the nest clause.
*/
func (this *Nest) Left() FromTerm {
	return this.left
}

/*
Returns the right term in the nest clause.
*/
func (this *Nest) Right() *KeyspaceTerm {
	return this.right
}

/*
Returns a boolean value depending on if it is
an outer or inner nest.
*/
func (this *Nest) Outer() bool {
	return this.outer
}

/*
Marshals input nest terms into byte array.
*/
func (this *Nest) MarshalJSON() ([]byte, error) {
	r := map[string]interface{}{"type": "nest"}
	r["left"] = this.left
	r["right"] = this.right
	r["outer"] = this.outer
	return json.Marshal(r)
}