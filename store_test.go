package main
import (
	"testing"
    "bytes"
)


func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc : DefaultPathTransformFunc,
	}
	s := NewStore(opts)

	data := bytes.NewReader([]byte("some jpeg bytes"))
	err := s.writeStream("mySpecialKey", data)
	if err != nil {
		t.Error(err)
	}

}