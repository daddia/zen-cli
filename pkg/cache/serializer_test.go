package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONSerializer(t *testing.T) {
	serializer := NewJSONSerializer[TestData]()

	testData := TestData{Name: "test", Value: 42}

	// Test Serialize
	data, err := serializer.Serialize(testData)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test Deserialize
	result, err := serializer.Deserialize(data)
	require.NoError(t, err)
	assert.Equal(t, testData.Name, result.Name)
	assert.Equal(t, testData.Value, result.Value)

	// Test ContentType
	assert.Equal(t, "application/json", serializer.ContentType())
}

func TestJSONSerializer_InvalidData(t *testing.T) {
	serializer := NewJSONSerializer[TestData]()

	// Test Deserialize with invalid JSON
	_, err := serializer.Deserialize([]byte("invalid json"))
	assert.Error(t, err)
}

func TestStringSerializer(t *testing.T) {
	serializer := NewStringSerializer()

	testString := "Hello, World!"

	// Test Serialize
	data, err := serializer.Serialize(testString)
	require.NoError(t, err)
	assert.Equal(t, []byte(testString), data)

	// Test Deserialize
	result, err := serializer.Deserialize(data)
	require.NoError(t, err)
	assert.Equal(t, testString, result)

	// Test ContentType
	assert.Equal(t, "text/plain", serializer.ContentType())
}

func TestStringSerializer_EmptyString(t *testing.T) {
	serializer := NewStringSerializer()

	// Test empty string
	data, err := serializer.Serialize("")
	require.NoError(t, err)
	assert.Equal(t, []byte{}, data)

	result, err := serializer.Deserialize(data)
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestJSONSerializer_ComplexData(t *testing.T) {
	serializer := NewJSONSerializer[map[string]interface{}]()

	testData := map[string]interface{}{
		"string": "test",
		"number": 42,
		"bool":   true,
		"array":  []string{"a", "b", "c"},
		"nested": map[string]string{
			"key": "value",
		},
	}

	// Test Serialize
	data, err := serializer.Serialize(testData)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test Deserialize
	result, err := serializer.Deserialize(data)
	require.NoError(t, err)
	assert.Equal(t, "test", result["string"])
	assert.Equal(t, float64(42), result["number"]) // JSON numbers become float64
	assert.Equal(t, true, result["bool"])
}

func TestJSONSerializer_NilData(t *testing.T) {
	serializer := NewJSONSerializer[*TestData]()

	// Test nil pointer
	data, err := serializer.Serialize(nil)
	require.NoError(t, err)

	result, err := serializer.Deserialize(data)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestSerializerInterfaces(t *testing.T) {
	// Test that our serializers implement the interface properly
	var _ Serializer[TestData] = NewJSONSerializer[TestData]()
	var _ Serializer[string] = NewStringSerializer()

	// Test content types
	jsonSer := NewJSONSerializer[TestData]()
	stringSer := NewStringSerializer()

	assert.Equal(t, "application/json", jsonSer.ContentType())
	assert.Equal(t, "text/plain", stringSer.ContentType())
}

func TestJSONSerializer_EmptyStruct(t *testing.T) {
	serializer := NewJSONSerializer[struct{}]()

	data, err := serializer.Serialize(struct{}{})
	require.NoError(t, err)
	assert.Equal(t, []byte("{}"), data)

	result, err := serializer.Deserialize(data)
	require.NoError(t, err)
	assert.Equal(t, struct{}{}, result)
}
