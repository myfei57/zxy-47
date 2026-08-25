package store

type MarkStore struct {
	st *Store
}

func NewMarks(st *Store) *MarkStore {
	return &MarkStore{st: st}
}

func (m *MarkStore) path(kind, key string) string {
	return "marks/" + kind + "/" + key + ".json"
}

func (m *MarkStore) Set(kind, key, payload string) error {
	return m.st.WriteJSON(m.path(kind, key), map[string]string{"key": key, "payload": payload})
}

func (m *MarkStore) Has(kind, key string) bool {
	return m.st.Exists(m.path(kind, key))
}

func (m *MarkStore) Payload(kind, key string) (string, error) {
	var doc map[string]string
	if err := m.st.ReadJSON(m.path(kind, key), &doc); err != nil {
		return "", err
	}
	return doc["payload"], nil
}

func (m *MarkStore) Clear(kind, key string) error {
	return m.st.Delete(m.path(kind, key))
}

func (m *MarkStore) Keys(kind string) ([]string, error) {
	return m.st.ListFiles("marks/" + kind)
}
