package formatters

import (
	"github.com/ugorji/go/codec"
	"io"
	"testing"
)
import (
	"encoding/json"
)

type HugeArrayElement struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
}

func generateHugeArray(size int) []HugeArrayElement {
	hugeArray := make([]HugeArrayElement, size)
	for i := 0; i < size; i++ {
		hugeArray[i] = HugeArrayElement{
			ID:    i,
			Value: "value",
		}
	}
	return hugeArray
}

func BenchmarkCodecStreamingEncode(b *testing.B) {
	b.ReportAllocs()

	const numElements = 100000

	var jh codec.JsonHandle
	enc := codec.NewEncoder(io.Discard, &jh)

	for i := 0; i < b.N; i++ {
		// Reset the encoder to avoid memory leaks
		enc.Reset(io.Discard)

		// Write the opening bracket for the array
		err := enc.Encode("[")
		if err != nil {
			panic(err)
		}

		for i := 0; i < numElements; i++ {
			element := HugeArrayElement{
				ID:    i,
				Value: "value",
			}

			// Encode each element in the array
			err := enc.Encode(element)
			if err != nil {
				panic(err)
			}

			// Write a comma between elements, except for the last element
			if i < numElements-1 {
				err = enc.Encode(",")
				if err != nil {
					panic(err)
				}
			}
		}
		// Write the closing bracket for the array
		err = enc.Encode("]")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	b.ReportAllocs()

	hugeArray := generateHugeArray(100000)

	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(hugeArray)
		if err != nil {
			b.Fatal(err)
		}
	}
}
