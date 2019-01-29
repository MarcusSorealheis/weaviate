/*                          _       _
 *__      _____  __ ___   ___  __ _| |_ ___
 *\ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
 * \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
 *  \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
 *
 * Copyright © 2016 - 2018 Weaviate. All rights reserved.
 * LICENSE: https://github.com/creativesoftwarefdn/weaviate/blob/develop/LICENSE.md
 * AUTHOR: Bob van Luijt (bob@kub.design)
 * See www.creativesoftwarefdn.org for details
 * Contact: @CreativeSofwFdn / bob@kub.design
 */
package common

import (
	"encoding/json"
	"fmt"

	"github.com/graphql-go/graphql"
)

// JSONNumberResolver turns json.Number types into number types usable by graphQL
func JSONNumberResolver(p graphql.ResolveParams) (interface{}, error) {
	sourceMap, ok := p.Source.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("source is not a map, but %t", p.Source)
	}

	field, ok := sourceMap[p.Info.FieldName]
	if !ok {
		return nil, fmt.Errorf("sourcemap has no field '%s', got %#v", p.Info.FieldName, sourceMap)
	}

	switch n := field.(type) {
	case json.Number:
		return n.Float64()
	case int64:
		return float64(n), nil
	case int:
		return float64(n), nil
	case float64:
		return n, nil
	}

	return nil, fmt.Errorf("unknown number type for %t", field)
}
