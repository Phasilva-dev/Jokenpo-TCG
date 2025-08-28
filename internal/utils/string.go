package utils

import (
	"strings"
	"fmt"
)

func SliceToString[T any](label string, items []T) string {
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("----- %s -----\n", label))

    if len(items) == 0 {
        sb.WriteString("(Empty)\n")
    } else {
        for i, item := range items {
            sb.WriteString(fmt.Sprintf("[%d]: %v\n", i, item))
        }
    }

    sb.WriteString("--------------------\n")
    return sb.String()
}