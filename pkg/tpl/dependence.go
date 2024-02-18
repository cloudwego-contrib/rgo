/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tpl

var BaseClientDependence = map[string]struct{}{
	"unsafe":                                    {},
	"github.com/cloudwego/kitex/client":         {},
	"github.com/cloudwego/kitex/pkg/rpcinfo":    {},
	"github.com/cloudwego/kitex/client/callopt": {},
}

var BaseSrvMainDependence = map[string]struct{}{
	"net":                               {},
	"github.com/cloudwego/kitex/server": {},
}

var BaseSrvHandlerDependence = map[string]struct{}{
	"unsafe": {},
}

func AddDependence(dependenceMap map[string]struct{}, dependence []string) {
	for _, d := range dependence {
		if _, ok := dependenceMap[d]; !ok {
			dependenceMap[d] = struct{}{}
		}
	}
}
