package x

import (
	"encoding/json"
	"fmt"
	"sync"
)

func MarshalIndentSyncMap(m *sync.Map) ([]byte, error) {
	tmp := make(map[string]any)

	m.Range(func(key, value any) bool {
		ks, ok := key.(string)
		if !ok {
			// Key types that are not strings must be handled explicitly,
			// because JSON object keys must be strings.
			ks = fmt.Sprint(key)
		}
		tmp[ks] = value
		return true
	})

	return json.MarshalIndent(tmp, "", "\t")
}
