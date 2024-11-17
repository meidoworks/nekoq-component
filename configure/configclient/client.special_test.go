package configclient_test

import (
	"testing"

	"github.com/meidoworks/nekoq-component/configure/configclient"
)

func TestEnvClientBasic(t *testing.T) {
	ec := configclient.NewEnvClient()

	nec := ec.CaseInsensitive()
	if nec.ParserWarning() != nil {
		t.Fatalf("parser warning: %s", nec.ParserWarning().Error())
	}

	p := configclient.Must(ec.CaseInsensitive().GetString("Path"))
	if p == "" {
		t.Fatal("PATH should not be empty")
	}
	t.Log("path=", p)
	nn := configclient.MustDefault("default_val").On(ec.GetString("HELLO_WORLD"))
	if nn == "default_val" {
		t.Log("hit default value")
	} else if nn == "yes" {
		t.Log("hit env setting")
	} else {
		t.Fatal("unexpected value from environment")
	}
}
