package store

type ParamDoc struct {
	Key   string         `json:"key"`
	Value map[string]any `json:"value"`
}

type ParamsStore struct {
	kv *KeyValue
}

func NewParams(st *Store) *ParamsStore {
	return &ParamsStore{kv: NewKeyValue(st)}
}

func (p *ParamsStore) Save(kind, key string, value map[string]any) error {
	return p.kv.Save("params/"+kind, key, ParamDoc{Key: key, Value: value})
}

func (p *ParamsStore) Load(kind, key string) (map[string]any, error) {
	var doc ParamDoc
	if err := p.kv.Load("params/"+kind, key, &doc); err != nil {
		return nil, err
	}
	return doc.Value, nil
}

func (p *ParamsStore) Exists(kind, key string) bool {
	return p.kv.Exists("params/"+kind, key)
}

func (p *ParamsStore) Delete(kind, key string) error {
	return p.kv.Delete("params/"+kind, key)
}
