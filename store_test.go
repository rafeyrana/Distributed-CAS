package main
import (
	"testing"
    "bytes"

)



func TestPathTransformFuncs(t *testing.T) {
	key := "momsbestpicture"
	
	pathKey := CASPathTransformFunc(key)
	expectedOriginalKey := "6804429f74181a63c50c3d81d733a12f14a353ff"
	expectedPath := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	if pathKey.Original != expectedOriginalKey || pathKey.PathName != expectedPath {
		t.Errorf("expected %s, got %s", expectedPath, pathKey.PathName)
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