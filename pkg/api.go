package pkg

import "fmt"

func formatPort(port int) string {
	return fmt.Sprintf(":%d", port)
}
