package configapi

import (
	"testing"
)

func TestSelectorCache(t *testing.T) {
	s := &Selectors{
		Data: map[string]string{
			"C": "C",
			"A": "A",
			"B": "B",
		},
	}
	s.cache()
	if s.cached != "A=A,B=B,C=C" {
		t.Fatal("cache field")
	}

	s = &Selectors{
		Data: map[string]string{
			"C": "C",
		},
	}
	s.cache()
	if s.cached != "C=C" {
		t.Fatal("cache field")
	}

	s = &Selectors{}
	s.cache()
	if s.cached != "" {
		t.Fatal("cache field")
	}
}

func TestSelectors_Fill(t *testing.T) {
	s := &Selectors{}
	if err := s.Fill("dc=1,env=2,app=good"); err != nil {
		t.Fatal(err)
	}
	if len(s.Data) != 3 {
		t.Fatal("3 items expected")
	}
	t.Log(s.Data)
}
