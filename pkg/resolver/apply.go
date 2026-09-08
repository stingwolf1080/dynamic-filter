package resolver

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Dependency applies resolver results to the same model pointer that will be
// persisted by the repository.
type Dependency struct{ model reflect.Value }

func NewDependency(model any) (*Dependency, error) {
	v := reflect.ValueOf(model)
	if !v.IsValid() || v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidResolvedData
	}
	return &Dependency{model: v}, nil
}

// Apply stores data in the field declared by ref's into option.
func (d *Dependency) Apply(ref Reference, data any) error {
	if ref.Target == "" || data == nil {
		return nil
	}
	field, err := d.targetField(ref.Target)
	if err != nil {
		return err
	}

	resolved := reflect.ValueOf(data)
	if resolved.IsValid() && resolved.Type().AssignableTo(field.Type()) {
		field.Set(resolved)
		return nil
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode resolved target %s: %w", ref.Target, err)
	}
	if err := json.Unmarshal(encoded, field.Addr().Interface()); err != nil {
		return fmt.Errorf("apply resolved target %s: %w", ref.Target, err)
	}
	return nil
}

// ValidateTarget verifies the into field before a resolver performs a query.
func (d *Dependency) ValidateTarget(ref Reference) error {
	if ref.Target == "" {
		return nil
	}
	_, err := d.targetField(ref.Target)
	return err
}

func (d *Dependency) targetField(target string) (reflect.Value, error) {
	if d == nil {
		return reflect.Value{}, ErrInvalidResolvedData
	}
	field := d.model.Elem().FieldByName(target)
	if !field.IsValid() || !field.CanSet() {
		return reflect.Value{}, fmt.Errorf("%w: %q", ErrReferenceTargetNotFound, target)
	}
	return field, nil
}

// ApplyResolvedField is a convenience wrapper for applying one resolved value.
func ApplyResolvedField(model any, target string, data any) error {
	dependency, err := NewDependency(model)
	if err != nil {
		return err
	}
	return dependency.Apply(Reference{Target: target}, data)
}
