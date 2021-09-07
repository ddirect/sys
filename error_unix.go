package sys

import (
	"fmt"
)

func unixError(api string, path string, err error) error {
	return fmt.Errorf("%s on '%s': %w", api, path, err)
}
