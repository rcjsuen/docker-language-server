package compose

import (
	"bytes"
	_ "embed"
	"slices"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

//go:embed compose-spec.json
var schemaData []byte

var composeSchema *jsonschema.Schema

func init() {
	schema, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaData))
	if err != nil {
		return
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", schema); err != nil {
		return
	}
	compiled, err := compiler.Compile("schema.json")
	if err != nil {
		return
	}
	composeSchema = compiled
}

func schemaProperties() map[string]*jsonschema.Schema {
	return composeSchema.Properties
}

func nodeProperties(nodes []*yaml.Node, line, column int) map[string]*jsonschema.Schema {
	if composeSchema != nil && slices.Contains(composeSchema.Types.ToStrings(), "object") && composeSchema.Properties != nil {
		if prop, ok := composeSchema.Properties[nodes[0].Value]; ok {
			for regexp, property := range prop.PatternProperties {
				if regexp.MatchString(nodes[1].Value) {
					if property.Ref != nil {
						return recurseNodeProperties(nodes, line, column, 2, property.Ref.Properties)
					}
				}
			}
		}
	}
	return nil
}

func recurseNodeProperties(nodes []*yaml.Node, line, column, nodeOffset int, properties map[string]*jsonschema.Schema) map[string]*jsonschema.Schema {
	if len(nodes) == nodeOffset || (len(nodes) >= nodeOffset+2 && nodes[nodeOffset].Column <= column && column < nodes[nodeOffset+1].Column) {
		return properties
	} else if column == nodes[nodeOffset].Column {
		return properties
	}

	value := nodes[nodeOffset].Value
	if prop, ok := properties[value]; ok {
		if prop.Ref != nil {
			if len(prop.Ref.Properties) > 0 {
				return recurseNodeProperties(nodes, line, column, nodeOffset+1, prop.Ref.Properties)
			}
			for regexp, property := range prop.Ref.PatternProperties {
				if regexp.MatchString(nodes[nodeOffset+1].Value) {
					for _, nested := range property.OneOf {
						if slices.Contains(nested.Types.ToStrings(), "object") {
							return recurseNodeProperties(nodes, line, column, nodeOffset+2, nested.Properties)
						}
					}
				}
			}
		}

		for _, schema := range prop.OneOf {
			if schema.Types != nil && slices.Contains(schema.Types.ToStrings(), "object") {
				if len(schema.Properties) > 0 {
					return recurseNodeProperties(nodes, line, column, nodeOffset+1, schema.Properties)
				}

				for regexp, property := range schema.PatternProperties {
					if len(nodes) == nodeOffset+1 {
						return nil
					}

					if regexp.MatchString(nodes[nodeOffset+1].Value) {
						for _, nested := range property.OneOf {
							if slices.Contains(nested.Types.ToStrings(), "object") {
								return recurseNodeProperties(nodes, line, column, nodeOffset+2, nested.Properties)
							}
						}
					}
				}
			}
		}

		if schema, ok := prop.Items.(*jsonschema.Schema); ok {
			for _, nested := range schema.OneOf {
				if nested.Types != nil && slices.Contains(nested.Types.ToStrings(), "object") {
					if len(nested.Properties) > 0 {
						return recurseNodeProperties(nodes, line, column, nodeOffset+1, nested.Properties)
					}
				}
			}
		}

		if nodes[nodeOffset].Column < column {
			if nodes[nodeOffset].Line == line {
				if prop.Enum == nil {
					return nil
				}
				enumSchema := &jsonschema.Schema{Types: prop.Types}
				enumValues := make(map[string]*jsonschema.Schema)
				for _, value := range prop.Enum.Values {
					enumValues[value.(string)] = enumSchema
				}
				return enumValues
			}
			return recurseNodeProperties(nodes, line, column, nodeOffset+1, prop.Properties)
		}
		return prop.Properties
	}
	return nil
}
