package models

// Translation is one field, of one row, in one language.
type Translation struct {
	Entity   string `json:"entity"`
	EntityID int64  `json:"entityId"`
	Field    string `json:"field"`
	Locale   string `json:"locale"`
	Value    string `json:"value"`
}

// TranslationSet is every translation for a single locale, indexed for lookup
// while walking the content payload: entity → row id → field → text.
//
// Settings have no row id and use 0, with the setting key as the field.
type TranslationSet map[string]map[int64]map[string]string

// Text returns the translation for a field, and whether one exists. A missing
// entry is normal and means "keep the English", not an error.
func (s TranslationSet) Text(entity string, id int64, field string) (string, bool) {
	rows, ok := s[entity]
	if !ok {
		return "", false
	}
	fields, ok := rows[id]
	if !ok {
		return "", false
	}
	value, ok := fields[field]
	return value, ok
}

// Put records one translation, creating the intermediate maps as needed.
func (s TranslationSet) Put(entity string, id int64, field, value string) {
	rows, ok := s[entity]
	if !ok {
		rows = map[int64]map[string]string{}
		s[entity] = rows
	}
	fields, ok := rows[id]
	if !ok {
		fields = map[string]string{}
		rows[id] = fields
	}
	fields[field] = value
}

// Fields returns every translated field for one row, for the admin panel to
// render beside the English. The result is never nil, so a caller can range
// over it without checking.
func (s TranslationSet) Fields(entity string, id int64) map[string]string {
	if rows, ok := s[entity]; ok {
		if fields, ok := rows[id]; ok {
			return fields
		}
	}
	return map[string]string{}
}
