package utils

import (
	"sort"

	"github.com/iancoleman/orderedmap"
)

// SortOrderedMapRecursively takes any value, recursively finds OrderedMaps,
// and sorts their keys for stable output
func SortOrderedMapRecursively(v any) any {
	switch val := v.(type) {
	case *orderedmap.OrderedMap:
		// Sort the keys of the pointer to OrderedMap
		val.SortKeys(sort.Strings)

		// Recursively process values
		for _, key := range val.Keys() {
			subVal, _ := val.Get(key)
			sortedSubVal := SortOrderedMapRecursively(subVal)
			val.Set(key, sortedSubVal)
		}
		return val

	case orderedmap.OrderedMap:
		// Create a new OrderedMap with sorted keys
		result := orderedmap.New()

		// Copy all key-values, sorting nested items
		for _, key := range val.Keys() {
			subVal, _ := val.Get(key)
			sortedSubVal := SortOrderedMapRecursively(subVal)
			result.Set(key, sortedSubVal)
		}

		result.SortKeys(sort.Strings)
		return result

	// case map[string]any:
	// 	// Convert to OrderedMap and sort
	// 	result := orderedmap.New()
	// 	for key, subVal := range val {
	// 		sortedSubVal := SortOrderedMapRecursively(subVal)
	// 		result.Set(key, sortedSubVal)
	// 	}
	// 	result.SortKeys(sort.Strings)
	// 	return result

	case []any:
		// Process each item in the slice
		for i, item := range val {
			val[i] = SortOrderedMapRecursively(item)
		}
		return val

	default:
		// Return as is for other types
		return v
	}
}
