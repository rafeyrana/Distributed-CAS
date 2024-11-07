package main
import (
	"testing"
    "bytes"

)



func TestPathTransformFuncs(t *testing.T) {
	key := "momsbestpicture"
	pathName := CASPathTransformFunc(key)
	expectedPath := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	if pathName != expectedPath {
		t.Errorf("expected %s, got %s", expectedPath, pathName)
	}
	
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc : CASPathTransformFunc,
	}
	s := NewStore(opts)

	data := bytes.NewReader([]byte("some jpeg bytes"))
	err := s.writeStream("momsbestpicture", data)
	if err != nil {
		t.Error(err)
	}

}