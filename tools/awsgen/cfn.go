package main

import (
	"encoding/json"
	"fmt"
)

// The subset of a CloudFormation resource type schema that the code
// generator consumes.
type cfnSchema struct {
	TypeName             string                  `json:"typeName"`
	Description          string                  `json:"description"`
	Definitions          map[string]*cfnProperty `json:"definitions"`
	Properties           map[string]*cfnProperty `json:"properties"`
	Required             []string                `json:"required"`
	PrimaryIdentifier    []string                `json:"primaryIdentifier"`
	ReadOnlyProperties   []string                `json:"readOnlyProperties"`
	CreateOnlyProperties []string                `json:"createOnlyProperties"`
	WriteOnlyProperties  []string                `json:"writeOnlyProperties"`
	Tagging              *cfnTagging             `json:"tagging"`
}

type cfnTagging struct {
	Taggable    bool   `json:"taggable"`
	TagProperty string `json:"tagProperty"`
}

// A node in a CloudFormation JSON Schema. Type may be a single
// string or a list of strings, so it is decoded via stringList.
type cfnProperty struct {
	Type                 stringList              `json:"type"`
	Ref                  string                  `json:"$ref"`
	Description          string                  `json:"description"`
	Properties           map[string]*cfnProperty `json:"properties"`
	Items                *cfnProperty            `json:"items"`
	Required             []string                `json:"required"`
	OneOf                []*cfnProperty          `json:"oneOf"`
	AnyOf                []*cfnProperty          `json:"anyOf"`
	AllOf                []*cfnProperty          `json:"allOf"`
	Enum                 []json.RawMessage       `json:"enum"`
	Pattern              string                  `json:"pattern"`
	Minimum              *float64                `json:"minimum"`
	Maximum              *float64                `json:"maximum"`
	MinLength            *int                    `json:"minLength"`
	MaxLength            *int                    `json:"maxLength"`
	Default              json.RawMessage         `json:"default"`
	AdditionalProperties json.RawMessage         `json:"additionalProperties"`
	PatternProperties    map[string]*cfnProperty `json:"patternProperties"`
}

// Decodes a JSON value that may be a single string or an array of
// strings into a slice.
type stringList []string

func (s *stringList) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	if data[0] == '[' {
		var list []string
		if err := json.Unmarshal(data, &list); err != nil {
			return err
		}
		*s = list
		return nil
	}

	var single string
	if err := json.Unmarshal(data, &single); err != nil {
		return err
	}

	*s = []string{single}
	return nil
}

func loadCFNSchema(data []byte) (*cfnSchema, error) {
	var schema cfnSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parsing CloudFormation schema: %w", err)
	}
	if schema.TypeName == "" {
		return nil, fmt.Errorf("CloudFormation schema is missing typeName")
	}
	return &schema, nil
}
