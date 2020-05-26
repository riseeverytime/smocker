package types

import (
	"encoding/json"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestStringMatcher_JSON(t *testing.T) {
	test := `"test"`

	var res StringMatcher
	if err := json.Unmarshal([]byte(test), &res); err != nil {
		t.Fatal(err)
	}

	if res.Matcher != DefaultMatcher {
		t.Fatalf("matcher %s should be equal to %s", res.Matcher, DefaultMatcher)
	}
	if res.Value != "test" {
		t.Fatalf("value %s should be equal to %s", res.Value, "test")
	}

	b, err := json.Marshal(&res)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != test {
		t.Fatalf("serialized value %s should be equal to %s", string(b), test)
	}

	test = `{"matcher":"test","value":"test2"}`
	res = StringMatcher{}
	if err = json.Unmarshal([]byte(test), &res); err != nil {
		t.Fatal(err)
	}

	if res.Matcher != "test" {
		t.Fatalf("matcher %s should be equal to %s", res.Matcher, "test")
	}
	if res.Value != "test2" {
		t.Fatalf("value %s should be equal to %s", res.Value, "test2")
	}

	b, err = json.Marshal(&res)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != test {
		t.Fatalf("serialized value %s should be equal to %s", string(b), test)
	}
}

func TestStringMatcher_YAML(t *testing.T) {
	test := `test`
	var res StringMatcher
	if err := yaml.Unmarshal([]byte(test), &res); err != nil {
		t.Fatal(err)
	}

	if res.Matcher != DefaultMatcher {
		t.Fatalf("matcher %s should be equal to %s", res.Matcher, DefaultMatcher)
	}
	if res.Value != "test" {
		t.Fatalf("value %s should be equal to %s", res.Value, "test")
	}

	b, err := yaml.Marshal(&res)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))

	test = `{"matcher":"test","value":"test2"}`
	res = StringMatcher{}
	if err = yaml.Unmarshal([]byte(test), &res); err != nil {
		t.Fatal(err)
	}

	if res.Matcher != "test" {
		t.Fatalf("matcher %s should be equal to %s", res.Matcher, "test")
	}
	if res.Value != "test2" {
		t.Fatalf("value %s should be equal to %s", res.Value, "test2")
	}

	b, err = yaml.Marshal(&res)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(b))
}

func TestMultiMapMatcher_JSON(t *testing.T) {
	test := `{"test":["test"]}`
	var res MultiMapMatcher
	if err := json.Unmarshal([]byte(test), &res); err != nil {
		t.Fatal(err)
	}

	if res == nil {
		t.Fatal("multimap matcher should not be nil")
	}

	for _, value := range res {
		if value.Matcher != DefaultMatcher {
			t.Fatalf("matcher %s should be equal to %s", value.Matcher, DefaultMatcher)
		}
	}

	expected := MultiMapMatcher{
		"test": {
			Matcher: "ShouldEqual",
			Value:   []string{"test"},
		},
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("values %v should be equal to %v", res, expected)
	}

	b, err := json.Marshal(&res)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != test {
		t.Fatalf("serialized value %s should be equal to %s", string(b), test)
	}

	test = `{"test":{"matcher":"test2","value":["test3"]}}`
	res = MultiMapMatcher{}
	if err = json.Unmarshal([]byte(test), &res); err != nil {
		t.Fatal(err)
	}

	if res["test"].Matcher != "test2" {
		t.Fatalf("matcher %s should be equal to %s", res["test"].Matcher, "test")
	}

	expected = MultiMapMatcher{
		"test": {
			Matcher: "test2",
			Value:   []string{"test3"},
		},
	}

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("values %v should be equal to %v", res, expected)
	}

	b, err = json.Marshal(&res)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != test {
		t.Fatalf("serialized value %s should be equal to %s", string(b), test)
	}
}
