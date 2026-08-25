package store

import "path/filepath"

type KeyValue struct {
	st *Store
}

func NewKeyValue(st *Store) *KeyValue {
	return &KeyValue{st: st}
}

func (k *KeyValue) Store() *Store {
	return k.st
}

func (k *KeyValue) Save(kind, key string, v any) error {
	return k.st.WriteJSON(filepath.ToSlash(filepath.Join(kind, key+".json")), v)
}

func (k *KeyValue) Load(kind, key string, v any) error {
	return k.st.ReadJSON(filepath.ToSlash(filepath.Join(kind, key+".json")), v)
}

func (k *KeyValue) Delete(kind, key string) error {
	return k.st.Delete(filepath.ToSlash(filepath.Join(kind, key+".json")))
}

func (k *KeyValue) Exists(kind, key string) bool {
	return k.st.Exists(filepath.ToSlash(filepath.Join(kind, key+".json")))
}

func (k *KeyValue) List(kind string) ([]string, error) {
	return k.st.ListFiles(kind)
}
