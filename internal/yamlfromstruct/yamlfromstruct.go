package yamlfromstruct

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"go.uber.org/zap"
	yaml2 "sigs.k8s.io/yaml"

	"github.com/yanakipre/bot/internal/encodingtooling"
	"github.com/yanakipre/bot/internal/logger"
)

// Regexp definitions
var keyMatchRegex = regexp.MustCompile(`"(\w+)":`)

type conventionalMarshaller struct {
	Value any
}

func (c conventionalMarshaller) MarshalJSON() ([]byte, error) {
	marshalled, err := json.Marshal(c.Value)
	if err != nil {
		return nil, err
	}

	converted := keyMatchRegex.ReplaceAllFunc(
		marshalled,
		func(match []byte) []byte {
			// Empty keys are valid JSON, only lowercase if we do not have an
			// empty key.
			if len(match) > 2 {
				return []byte(
					fmt.Sprintf(
						`"%s":`,
						encodingtooling.CamelToSnake(string(match[1:len(match)-2])),
					),
				)
			}
			return match
		},
	)

	return converted, err
}

// Generate returns YAML string, serializing structure, not requiring "yaml" tag on its fields.
func Generate(ctx context.Context, s any) string {
	bytes, err := json.MarshalIndent(conventionalMarshaller{s}, "", " ")
	if err != nil {
		logger.Fatal(ctx, "failed to marshal request to json", zap.Error(err))
	}
	toYAML, err := yaml2.JSONToYAML(bytes)
	if err != nil {
		logger.Fatal(ctx, "failed to marshal json to yaml", zap.Error(err))
	}
	return string(toYAML)
}
