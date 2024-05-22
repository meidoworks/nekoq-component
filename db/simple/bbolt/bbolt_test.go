package bbolt

import (
	"encoding/json"
	"testing"
)

type User struct {
	IdData string `json:"id"`
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Addr   string `json:"addr"`
}

func (u *User) Id() []byte {
	return []byte(u.IdData)
}

func (u *User) Marshal() ([]byte, error) {
	return json.Marshal(u)
}

func (u *User) Unmarshal(data []byte) error {
	return json.Unmarshal(data, u)
}

func TestBboltStoreOperations(t *testing.T) {
	s, err := NewBboltStore(&BboltStoreConfig{
		Path: "data.db",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func(s *BboltStore) {
		_ = s.Close()
	}(s)

	const table = "users"
	user := &User{
		IdData: "id1",
		Name:   "zhangsan",
		Age:    20,
		Addr:   "hello world",
	}

	if obj, err := s.Table(table).QueryById(user.Id(), new(User)); err != nil {
		t.Fatal(err)
	} else if obj == nil {
		t.Log("query by id is nil")
	} else {
		t.Log("query by id:", obj)
	}

	if err := s.Table(table).Insert(user); err != nil {
		t.Fatal(err)
	} else {
		t.Log("insert user")
	}

	if obj, err := s.Table(table).QueryById(user.Id(), new(User)); err != nil {
		t.Fatal(err)
	} else if obj == nil {
		t.Log("query by id is nil")
	} else {
		t.Log("query by id:", obj)
	}

	if err := s.Table(table).Delete(user.Id()); err != nil {
		t.Fatal(err)
	} else {
		t.Log("delete user")
	}

	if obj, err := s.Table(table).QueryById(user.Id(), new(User)); err != nil {
		t.Fatal(err)
	} else if obj == nil {
		t.Log("query by id is nil")
	} else {
		t.Log("query by id:", obj)
	}

	user.Age = 30
	if err := s.Table(table).Update(user); err != nil {
		t.Fatal(err)
	} else {
		t.Log("update user")
	}

	if obj, err := s.Table(table).QueryById(user.Id(), new(User)); err != nil {
		t.Fatal(err)
	} else if obj == nil {
		t.Log("query by id is nil")
	} else {
		t.Log("query by id:", obj)
	}

	if err := s.Table(table).Insert(user); err != nil {
		t.Fatal(err)
	} else {
		t.Log("insert user")
	}

	if obj, err := s.Table(table).QueryById(user.Id(), new(User)); err != nil {
		t.Fatal(err)
	} else if obj == nil {
		t.Log("query by id is nil")
	} else {
		t.Log("query by id:", obj)
	}

	user.Age = 40
	if err := s.Table(table).Update(user); err != nil {
		t.Fatal(err)
	} else {
		t.Log("update user")
	}

	if obj, err := s.Table(table).QueryById(user.Id(), new(User)); err != nil {
		t.Fatal(err)
	} else if obj == nil {
		t.Log("query by id is nil")
	} else {
		t.Log("query by id:", obj)
	}

	if err := s.Table(table).Insert(user); err != nil {
		t.Log("insert user duplicated")
	} else {
		t.Fatal("insert user should report duplication")
	}

	if obj, err := s.Table(table).QueryById(user.Id(), new(User)); err != nil {
		t.Fatal(err)
	} else if obj == nil {
		t.Log("query by id is nil")
	} else {
		t.Log("query by id:", obj)
	}

	if err := s.Table(table).Delete(user.Id()); err != nil {
		t.Fatal(err)
	} else {
		t.Log("delete user")
	}

}
