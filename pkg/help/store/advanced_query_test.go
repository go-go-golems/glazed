package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"github.com/go-go-golems/glazed/pkg/help/model"
)

const adv1 = `---
slug: adv-1
title: Advanced 1
sectionType: Example
flags:
  - flagA
commands:
  - cmdA
topics:
  - topicA
---
Content adv 1.`
const adv2 = `---
slug: adv-2
title: Advanced 2
sectionType: Tutorial
flags:
  - flagB
commands:
  - cmdB
topics:
  - topicB
---
Content adv 2.`
const adv3 = `---
slug: adv-3
title: Advanced 3
sectionType: Example
flags:
  - flagA
  - flagB
commands:
  - cmdA
  - cmdB
topics:
  - topicA
  - topicB
---
Content adv 3.`

func TestAdvancedQueries(t *testing.T) {
	dir := "test_adv_dir"
	os.Mkdir(dir, 0755)
	defer os.RemoveAll(dir)
	os.WriteFile(filepath.Join(dir, "adv1.md"), []byte(adv1), 0644)
	os.WriteFile(filepath.Join(dir, "adv2.md"), []byte(adv2), 0644)
	os.WriteFile(filepath.Join(dir, "adv3.md"), []byte(adv3), 0644)

	s, err := Open("test_adv.db")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() { s.Close(); os.Remove("test_adv.db") }()

	_, err = s.SyncMarkdownDir(dir)
	if err != nil {
		t.Fatalf("sync dir: %v", err)
	}

	ctx := context.Background()

	// Query: Example AND flagA
	q1 := query.And(query.IsType(model.SectionExample), query.HasFlag("flagA"))
	results, err := s.Find(ctx, q1)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for Example AND flagA, got %d", len(results))
	}

	// Query: (flagA OR flagB) AND commandB
	q2 := query.And(query.Or(query.HasFlag("flagA"), query.HasFlag("flagB")), query.HasCommand("cmdB"))
	c := &query.Compiler{}
	q2(c)
	sqlStr, args := c.SQL()
	t.Logf("Advanced Query SQL: %s, args: %v", sqlStr, args)
	results, err = s.Find(ctx, q2)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for (flagA OR flagB) AND cmdB, got %d", len(results))
	}

	// Query: NOT flagA
	q3 := query.NotHasFlag("flagA")
	results, err = s.Find(ctx, q3)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if len(results) != 1 || results[0].Slug != "adv-2" {
		t.Errorf("expected only adv-2 for NOT flagA, got %+v", results)
	}

	// Query: topicA AND topicB
	q4 := query.And(query.HasTopic("topicA"), query.HasTopic("topicB"))
	results, err = s.Find(ctx, q4)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if len(results) != 1 || results[0].Slug != "adv-3" {
		t.Errorf("expected only adv-3 for topicA AND topicB, got %+v", results)
	}

	// Query: NOT topicA
	q5 := query.NotHasTopic("topicA")
	results, err = s.Find(ctx, q5)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if len(results) != 1 || results[0].Slug != "adv-2" {
		t.Errorf("expected only adv-2 for NOT topicA, got %+v", results)
	}

	// Query: NOT cmdA
	q6 := query.NotHasCommand("cmdA")
	results, err = s.Find(ctx, q6)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if len(results) != 1 || results[0].Slug != "adv-2" {
		t.Errorf("expected only adv-2 for NOT cmdA, got %+v", results)
	}
} 