package cache

import (
	"encoding/json"
)

// Serializer handles serialization/deserialization of cached data
type Serializer[T any] interface {
	// Serialize converts data to bytes for storage
	Serialize(data T) ([]byte, error)

	// Deserialize converts bytes back to data
	Deserialize(data []byte) (T, error)

	// ContentType returns the content type for the serialized data
	ContentType() string
}

// JSONSerializer implements Serializer using JSON encoding
type JSONSerializer[T any] struct{}

// NewJSONSerializer creates a new JSON serializer
func NewJSONSerializer[T any]() *JSONSerializer[T] {
	return &JSONSerializer[T]{}
}

// Serialize converts data to JSON bytes
func (s *JSONSerializer[T]) Serialize(data T) ([]byte, error) {
	return json.Marshal(data)
}

// Deserialize converts JSON bytes back to data
func (s *JSONSerializer[T]) Deserialize(data []byte) (T, error) {
	var result T
	err := json.Unmarshal(data, &result)
	return result, err
}

// ContentType returns the JSON content type
func (s *JSONSerializer[T]) ContentType() string {
	return "application/json"
}

// StringSerializer implements Serializer for string data (no encoding needed)
type StringSerializer struct{}

// NewStringSerializer creates a new string serializer
func NewStringSerializer() *StringSerializer {
	return &StringSerializer{}
}

// Serialize converts string to bytes
func (s *StringSerializer) Serialize(data string) ([]byte, error) {
	return []byte(data), nil
}

// Deserialize converts bytes to string
func (s *StringSerializer) Deserialize(data []byte) (string, error) {
	return string(data), nil
}

// ContentType returns the plain text content type
func (s *StringSerializer) ContentType() string {
	return "text/plain"
}
