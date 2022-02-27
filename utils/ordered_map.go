// Package utils provides various utilities.
package utils

import (
	"encoding/json"
	"fmt"
)

// TODO: Make OrderMap generic once Go 1.18 is used

type OrderedMap struct {
	values map[string]interface{}
	keys   []string
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		values: make(map[string]interface{}),
	}
}

func (o *OrderedMap) Set(key string, value interface{}) {
	if _, ok := o.values[key]; !ok {
		o.keys = append(o.keys, key)
	}

	o.values[key] = value
}

func (o *OrderedMap) Remove(key string) {
	delete(o.values, key)

	index := -1
	for i, aKey := range o.keys {
		if aKey == key {
			index = i
			break
		}
	}
	if index != -1 {
		o.keys = append(o.keys[:index], o.keys[index+1:]...)
	}
}

func (o *OrderedMap) MarshalJSON() ([]byte, error) {
	result := "{"

	for i, key := range o.keys {
		if i > 0 {
			result += ","
		}

		marshaledKey, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		value, err := json.Marshal(o.values[key])
		if err != nil {
			return nil, err
		}

		result += fmt.Sprintf("%s:%s", marshaledKey, value)
	}

	result += "}"

	return []byte(result), nil
}

// ForEach calls callback for every item in the OrderedMap, in order
func (o *OrderedMap) ForEach(callback func(key string, value interface{})) {
	for _, key := range o.keys {
		callback(key, o.values[key])
	}
}
