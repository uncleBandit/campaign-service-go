// internal/service/template_service.go
package service

import (
    "strings"
)

func RenderTemplate(template string, data map[string]string) string {
    result := template
    for k, v := range data {
        result = strings.ReplaceAll(result, "{"+k+"}", v)
    }
    return result
}


