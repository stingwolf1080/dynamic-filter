package mongox

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type ModelMeta struct {
	Name     string             `json:"name"`
	RoleName string             `json:"role_name"`
	Group    string             `json:"group"`
	Table    string             `json:"-"`
	Index    []mongo.IndexModel `json:"-"`
	Model    any                `json:"-"`
	Perm     bool               `json:"-"`
}

func GetAllModels() []ModelMeta {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]ModelMeta, 0, len(models))
	for _, m := range models {
		out = append(out, m)
	}
	return out
}

func GetModels() map[string]ModelMeta {

	mu.RLock()
	defer mu.RUnlock()
	return models
}
