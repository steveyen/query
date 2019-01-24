//  Copyright (c) 2019 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

// +build enterprise

package inline

import (
	"encoding/json"
	go_errors "errors"
	"fmt"

	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/expression/parser"
	"github.com/couchbase/query/functions"
	"github.com/couchbase/query/value"
)

type inline struct {
}

type inlineBody struct {
	expr     expression.Expression
	varNames []string
}

func init() {
	functions.FunctionsNewLanguage(functions.INLINE, &inline{})
}

func (this *inline) Execute(name functions.FunctionName, body functions.FunctionBody, values []value.Value, context functions.Context) (value.Value, errors.Error) {
	var parent map[string]interface{}

	funcBody, ok := body.(*inlineBody)

	if !ok {
		return nil, errors.NewInternalFunctionError("Wrong language being executed!", name.Name())
	}

	if len(funcBody.varNames) != 0 {
		if len(values) != len(funcBody.varNames) {
			return nil, errors.NewArgumentsMismatchError(name.Name())
		}
		args := make([]value.Value, len(values))
		for i, _ := range values {
			args[i] = value.NewValue(values[i])
		}
		parent = map[string]interface{}{"args": args}
	} else {
		parent := make(map[string]interface{}, len(values))
		for i, _ := range values {
			parent[funcBody.varNames[i]] = values[i]
		}
	}
	val, err := funcBody.expr.Evaluate(value.NewValue(parent), context)
	if err != nil {
		return nil, errors.NewError(err, fmt.Sprintf("Error executing function %v", name.Name()))
	} else {
		return val, nil
	}
}

func NewInlineBody(expr expression.Expression, vars []string) (functions.FunctionBody, errors.Error) {
	return &inlineBody{expr: expr, varNames: vars}, nil
}

func (this *inlineBody) Lang() functions.Language {
	return functions.INLINE
}

func (this *inlineBody) Body(object map[string]interface{}) {
	object["language"] = "inline"
	object["expression"] = this.expr
	if len(this.varNames) == 0 {
		object["variadic"] = true
	} else {
		object["parameters"] = this.varNames
	}
}

func MakeInline(name functions.FunctionName, body []byte) (functions.FunctionBody, errors.Error) {
	var expr expression.Expression
	var _unmarshalled struct {
		_          string   `json:"#language"`
		Variadic   bool     `json:"variadic"`
		Parameters []string `json:"parameters"`
		Expression string   `json:"expression"`
	}
	err := json.Unmarshal(body, &_unmarshalled)
	if err != nil {
		return nil, errors.NewFunctionEncodingError("decode body", name.Name(), err)
	}
	if _unmarshalled.Expression != "" {
		expr, err = parser.Parse(_unmarshalled.Expression)
		if err != nil {
			return nil, errors.NewFunctionEncodingError("decode body", name.Name(), err)
		}
	} else {
		return nil, errors.NewFunctionEncodingError("decode body", name.Name(), go_errors.New("expression is missing"))
	}
	if len(_unmarshalled.Parameters) != 0 && _unmarshalled.Variadic {
		return nil, errors.NewFunctionEncodingError("decode body", name.Name(), go_errors.New("function body is variadic AND has parameter names"))
	}
	return &inlineBody{expr: expr, varNames: _unmarshalled.Parameters}, nil
}