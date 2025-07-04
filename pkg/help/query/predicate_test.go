package query

import (
	"testing"
	"github.com/go-go-golems/glazed/pkg/help/model"
)

func TestIsTypePredicate(t *testing.T) {
	c := &compiler{}
	IsType(model.SectionExample)(c)
	sql, args := c.SQL()
	if sql == "" || len(args) != 1 || args[0] != "Example" {
		t.Errorf("IsType: got sql=%q args=%v", sql, args)
	}
}

func TestHasTopicPredicate(t *testing.T) {
	c := &compiler{}
	HasTopic("foo")(c)
	sql, args := c.SQL()
	if sql == "" || len(args) != 1 || args[0] != "foo" {
		t.Errorf("HasTopic: got sql=%q args=%v", sql, args)
	}
}

func TestTextSearchPredicate(t *testing.T) {
	c := &compiler{}
	TextSearch("bar")(c)
	sql, args := c.SQL()
	if sql == "" || len(args) != 1 || args[0] != "bar" {
		t.Errorf("TextSearch: got sql=%q args=%v", sql, args)
	}
} 