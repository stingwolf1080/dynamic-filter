package resolver

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/stingwolf1080/dynamic-filter/pkg/cache"
)

// Reference describes one non-empty field marked with a ref tag.
type Reference struct {
	Field    string
	Name     string
	ID       string
	Required bool
	Target   string
}

type referenceMetadata struct {
	field    string
	name     string
	target   string
	index    []int
	required bool
}

type referenceMetadataResult struct {
	references []referenceMetadata
	err        error
}

// referenceMetadataCache stores immutable tag metadata by struct type. Runtime
// IDs are deliberately not cached: they are always read from the current model.
var referenceMetadataCache = cache.NewCache()

// ExtractReferences reads ref:"name[,required]" tags from a struct. Embedded structs
// are traversed, while unexported fields and unsupported field values are ignored.
func ExtractReferences(model any) ([]Reference, error) {
	v := reflect.ValueOf(model)
	for v.IsValid() && v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, ErrInvalidModel
		}
		v = v.Elem()
	}
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return nil, ErrInvalidModel
	}
	metadata, err := metadataFor(v.Type())
	if err != nil {
		return nil, err
	}

	refs := make([]Reference, 0, len(metadata))
	for _, item := range metadata {
		value, ok := valueByIndex(v, item.index)
		// A nil embedded pointer has no reference value to resolve, matching the
		// previous extractor behaviour.
		if !ok {
			continue
		}
		id, err := referenceID(value)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", item.field, err)
		}
		refs = append(refs, Reference{
			Field: item.field, Name: item.name, ID: id, Required: item.required, Target: item.target,
		})
	}
	return refs, nil
}

func metadataFor(modelType reflect.Type) ([]referenceMetadata, error) {
	key := modelType.PkgPath() + ":" + modelType.String()
	if cached, ok := referenceMetadataCache.Get(key); ok {
		result := cached.(referenceMetadataResult)
		return result.references, result.err
	}

	metadata, err := extractMetadata(modelType, modelType.Name(), modelType, nil)
	referenceMetadataCache.Set(key, referenceMetadataResult{references: metadata, err: err})
	return metadata, err
}

func extractMetadata(t reflect.Type, prefix string, modelType reflect.Type, parentIndex []int) ([]referenceMetadata, error) {
	refs := make([]referenceMetadata, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}
		index := append(append([]int(nil), parentIndex...), i)

		if field.Anonymous && field.Tag.Get("ref") == "" {
			nestedType := field.Type
			for nestedType.Kind() == reflect.Pointer {
				nestedType = nestedType.Elem()
			}
			if nestedType.Kind() == reflect.Struct {
				nested, err := extractMetadata(nestedType, fieldName, modelType, index)
				if err != nil {
					return nil, err
				}
				refs = append(refs, nested...)
			}
			continue
		}
		if field.PkgPath != "" {
			continue
		}
		name, required, target, ok, err := parseRefTag(field.Tag.Get("ref"))
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", fieldName, err)
		}
		if !ok {
			continue
		}
		if target != "" {
			if err := validateReferenceTarget(modelType, target); err != nil {
				return nil, fmt.Errorf("field %s: %w", fieldName, err)
			}
		}
		refs = append(refs, referenceMetadata{field: fieldName, name: name, target: target, index: index, required: required})
	}
	return refs, nil
}

func valueByIndex(value reflect.Value, index []int) (reflect.Value, bool) {
	for _, fieldIndex := range index {
		for value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return reflect.Value{}, false
			}
			value = value.Elem()
		}
		if value.Kind() != reflect.Struct {
			return reflect.Value{}, false
		}
		value = value.Field(fieldIndex)
	}
	return value, true
}

// validateReferenceTarget runs during tag extraction so a bad into target fails
// before a resolver can issue its database or remote-service query.
func validateReferenceTarget(modelType reflect.Type, target string) error {
	field, ok := modelType.FieldByName(target)
	if !ok || field.PkgPath != "" {
		return fmt.Errorf("%w: %q", ErrReferenceTargetNotFound, target)
	}
	return nil
}

func parseRefTag(tag string) (name string, required bool, target string, ok bool, err error) {
	if tag == "" || tag == "-" {
		return "", false, "", false, nil
	}
	parts := strings.Split(tag, ",")
	name = strings.TrimSpace(parts[0])
	if name == "" {
		return "", false, "", false, fmt.Errorf("reference name is empty")
	}
	for _, option := range parts[1:] {
		option = strings.TrimSpace(option)
		switch option {
		case "required":
			required = true
		case "":
		case "into":
			return "", false, "", false, fmt.Errorf("reference target is empty")
		default:
			if !strings.HasPrefix(option, "into=") {
				return "", false, "", false, fmt.Errorf("unknown ref option %q", option)
			}
			target = strings.TrimSpace(strings.TrimPrefix(option, "into="))
			if target == "" {
				return "", false, "", false, fmt.Errorf("reference target is empty")
			}
		}
	}
	return name, required, target, true, nil
}

func referenceID(v reflect.Value) (string, error) {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.String {
		return "", fmt.Errorf("ref field must be a string")
	}
	return strings.TrimSpace(v.String()), nil
}
