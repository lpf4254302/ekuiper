// Copyright 2023 EMQ Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package function

import (
	"fmt"
	"reflect"

	"github.com/lf-edge/ekuiper/pkg/api"
	"github.com/lf-edge/ekuiper/pkg/ast"
	"github.com/lf-edge/ekuiper/pkg/cast"
)

func registerObjectFunc() {
	builtins["keys"] = builtinFunc{
		fType: ast.FuncTypeScalar,
		exec: func(ctx api.FunctionContext, args []interface{}) (interface{}, bool) {
			arg := args[0]
			if arg, ok := arg.(map[string]interface{}); ok {
				list := make([]string, 0, len(arg))
				for key := range arg {
					list = append(list, key)
				}
				return list, true
			}
			return fmt.Errorf("the argument should be map[string]interface{}"), false
		},
		val:   ValidateOneArg,
		check: returnNilIfHasAnyNil,
	}
	builtins["values"] = builtinFunc{
		fType: ast.FuncTypeScalar,
		exec: func(ctx api.FunctionContext, args []interface{}) (interface{}, bool) {
			arg := args[0]
			if arg, ok := arg.(map[string]interface{}); ok {
				list := make([]interface{}, 0, len(arg))
				for _, value := range arg {
					list = append(list, value)
				}
				return list, true
			}
			return fmt.Errorf("the argument should be map[string]interface{}"), false
		},
		val:   ValidateOneArg,
		check: returnNilIfHasAnyNil,
	}
	builtins["object"] = builtinFunc{
		fType: ast.FuncTypeScalar,
		exec: func(ctx api.FunctionContext, args []interface{}) (interface{}, bool) {
			keys, ok := args[0].([]interface{})
			if !ok {
				return fmt.Errorf("first argument should be []string"), false
			}
			values, ok := args[1].([]interface{})
			if !ok {
				return fmt.Errorf("second argument should be []interface{}"), false
			}
			if len(keys) != len(values) {
				return fmt.Errorf("the length of the arguments should be same"), false
			}
			if len(keys) == 0 {
				return nil, true
			}
			m := make(map[string]interface{}, len(keys))
			for i, k := range keys {
				key, ok := k.(string)
				if !ok {
					return fmt.Errorf("first argument should be []string"), false
				}
				m[key] = values[i]
			}
			return m, true
		},
		val: func(ctx api.FunctionContext, args []ast.Expr) error {
			return ValidateLen(2, len(args))
		},
		check: returnNilIfHasAnyNil,
	}
	builtins["zip"] = builtinFunc{
		fType: ast.FuncTypeScalar,
		exec: func(ctx api.FunctionContext, args []interface{}) (interface{}, bool) {
			lists, ok := args[0].([]interface{})
			if !ok {
				return fmt.Errorf("each argument should be [][2]interface{}"), false
			}
			if len(lists) == 0 {
				return nil, true
			}
			m := make(map[string]interface{}, len(lists))
			for _, item := range lists {
				a, ok := item.([]interface{})
				if !ok {
					return fmt.Errorf("each argument should be [][2]interface{}"), false
				}
				if len(a) != 2 {
					return fmt.Errorf("each argument should be [][2]interface{}"), false
				}
				key, ok := a[0].(string)
				if !ok {
					return fmt.Errorf("the first element in the list item should be string"), false
				}
				m[key] = a[1]
			}
			return m, true
		},
		val:   ValidateOneArg,
		check: returnNilIfHasAnyNil,
	}
	builtins["items"] = builtinFunc{
		fType: ast.FuncTypeScalar,
		exec: func(ctx api.FunctionContext, args []interface{}) (interface{}, bool) {
			m, ok := args[0].(map[string]interface{})
			if !ok {
				return fmt.Errorf("first argument should be map[string]interface{}"), false
			}
			if len(m) < 1 {
				return nil, true
			}
			list := make([]interface{}, 0, len(m))
			for k, v := range m {
				list = append(list, []interface{}{k, v})
			}
			return list, true
		},
		val:   ValidateOneArg,
		check: returnNilIfHasAnyNil,
	}
	builtins["object_concat"] = builtinFunc{
		fType: ast.FuncTypeScalar,
		exec: func(ctx api.FunctionContext, args []interface{}) (interface{}, bool) {
			res := make(map[string]interface{})
			for i, arg := range args {
				if arg == nil {
					continue
				}
				arg, ok := arg.(map[string]interface{})
				if !ok {
					return fmt.Errorf("the argument should be map[string]interface{}, got %v", args[i]), false
				}
				for k, v := range arg {
					res[k] = v
				}
			}
			return res, true
		},
		val: func(_ api.FunctionContext, args []ast.Expr) error {
			return ValidateAtLeast(2, len(args))
		},
	}
	builtins["erase"] = builtinFunc{
		fType: ast.FuncTypeScalar,
		exec: func(ctx api.FunctionContext, args []interface{}) (interface{}, bool) {
			contains := func(array []string, target string) bool {
				for _, v := range array {
					if target == v {
						return true
					}
				}
				return false
			}
			if len(args) != 2 {
				return fmt.Errorf("the argument number should be 2, got %v", len(args)), false
			}
			res := make(map[string]interface{})
			argMap, ok := args[0].(map[string]interface{})
			if !ok {
				return fmt.Errorf("the first argument should be map[string]interface{}, got %v", args[0]), false
			}
			eraseArray := make([]string, 0)
			v := reflect.ValueOf(args[1])
			switch v.Kind() {
			case reflect.Slice:
				array, err := cast.ToStringSlice(args[1], cast.CONVERT_ALL)
				if err != nil {
					return err, false
				}
				eraseArray = append(eraseArray, array...)
			case reflect.String:
				str := args[1].(string)
				for k, v := range argMap {
					if k != str {
						res[k] = v
					}
				}
				return res, true
			default:
				return fmt.Errorf("the augument should be slice or string"), false
			}
			for k, v := range argMap {
				if !contains(eraseArray, k) {
					res[k] = v
				}
			}

			return res, true
		},
		val: func(_ api.FunctionContext, args []ast.Expr) error {
			return ValidateAtLeast(2, len(args))
		},
		check: returnNilIfHasAnyNil,
	}
}
