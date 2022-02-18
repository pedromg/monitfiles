package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewStore(t *testing.T) {

}

func TestValidPath(t *testing.T) {
	type Case struct {
		ID    int
		Path  string
		Valid bool
	}

	cases := []Case{
		{1, "", false},
		{2, ".", true},
		{3, "..", true},
		{4, "/", true},
		{5, "/tmp", true},
		{6, "./main_test.go", false},
	}

	for _, c := range cases {
		res, err := validPath(c.Path)
		if c.Valid {
			t.Logf("case %d result is %s", c.ID, res)
		}
		if (err == nil) != c.Valid {
			t.Errorf("case %d got %t should be %t", c.ID, (err == nil), c.Valid)
		}

	}
}

func TestValidFileTypes(t *testing.T) {
	type Case struct {
		ID    int
		List  string
		Res   []string
		Valid bool
	}

	cases := []Case{
		{1, "", []string{}, false},
		{2, "go", []string{"go"}, true},
		{3, "go rb", []string{"go", "rb"}, true},
		{4, "Go Rb", []string{"go", "rb"}, true},
		{5, "Go Rb", []string{"go", "rb"}, true},
		{6, "Go Rb Css html", []string{"go", "rb", "css", "html"}, false},
		{7, "Go Rb Css html", []string{"css", "go", "html", "rb"}, true},
		{8, "Go - .", []string{"-", ".", "go"}, true},
	}

	for _, c := range cases {
		res, err := validFileTypes(c.List)
		if c.Valid {
			t.Logf("case %d result is %v", c.ID, res)
		}
		if (err != nil) && c.Valid {
			t.Errorf("case %d got %t should be %t", c.ID, (err == nil), c.Valid)
		}
		if c.Valid && !cmp.Equal(res, c.Res) {
			t.Errorf("case %d diffs: \n%s", c.ID, cmp.Diff(res, c.Res))
		}
	}
}

func TestValidScript(t *testing.T) {
	type Case struct {
		ID     int
		Script string
		Valid  bool
	}

	cases := []Case{
		{1, "", false},
		{2, "go version", true},
	}

	for _, c := range cases {
		_, err := validScript(c.Script)
		if (err == nil) != c.Valid {
			t.Errorf("case %d got %t should be %t", c.ID, (err == nil), c.Valid)
		}
	}
}

// TODO: test main
