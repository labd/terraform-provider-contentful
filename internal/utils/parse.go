package utils

import (
	"fmt"
	"strings"
)

func ParseThreePartID(id string) ([]string, error) {
	idParts := strings.SplitN(id, ":", 3)
	if len(idParts) != 3 {
		return nil, fmt.Errorf("invalid ID format: %s, expected space_id:environment:resource_id", id)
	}
	return idParts, nil
}
