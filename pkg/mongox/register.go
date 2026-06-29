package mongox

import (
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/stingwolf1080/dynamic-filter/pkg/helper"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	models = map[string]ModelMeta{}
	mu     sync.RWMutex
)

func Register(model any, perm bool) {
	if model == nil {
		panic("mongox: nil model")
	}

	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.Name()

	mu.Lock()
	defer mu.Unlock()

	if _, exists := models[name]; exists {
		return
	}
	var indexes []mongo.IndexModel
	var err error
	if indexes, err = parseModel(model); err != nil {
		panic(err)
	}
	log.Printf("[debug][register model]: %v\n", name)
	models[name] = ModelMeta{
		Name:     name,
		RoleName: helper.GetRoleName(name),
		Table:    helper.GetModelTable(model),
		Group:    extractGroup(t.PkgPath()),
		Index:    indexes,
		Model:    model,
		Perm:     perm,
	}
}

func extractGroup(pkg string) string {
	parts := strings.Split(pkg, "/")

	for i := 0; i < len(parts)-2; i++ {
		if parts[i] == "module" && parts[i+2] == "model" {
			return parts[i+1]
		}
	}

	return "none"
}
