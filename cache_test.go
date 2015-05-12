package main

import (
	"testing"
)

func TestCache(t *testing.T) {
	initialize()

	files := []string{"file1", "file2", "file3"}

	for _, file := range files {
		err := cacheAdd(link, file)
		if err != nil {
			t.Errorf("Failed to add %s to the cache: %s", file, err)
		}
	}

	for _, file := range files {
		if b, err := cacheContains(link, file); !b || err != nil {
			t.Errorf("cache should contains %s", file)
		}
	}

	cacheRemove(link, "file1")

	if b, err := cacheContains(link, "file1"); b || err != nil {
		t.Error("cache should not contains file1")
	}
	if b, err := cacheContains(link, "file2"); !b || err != nil {
		t.Error("cache should still contains file2")
	}

	invalideCache()

	for _, file := range files {
		if b, err := cacheContains(link, file); b || err != nil {
			t.Errorf("cache should be empty, but contains %s", file)
		}
	}

}

func TestLoadCache(t *testing.T) {
	initialize()

	files := []string{"file1", "file2", "file3"}

	for _, file := range files {
		err := cacheAdd(link, file)
		if err != nil {
			t.Errorf("Failed to add %s to the cache", file)
		}
	}

	cache = Cache{}

	for _, file := range files {
		if b, err := cacheContains(link, file); b || err != nil {
			t.Errorf("cache should be empty, but contains %s", file)
		}
	}

	loadCache()

	for _, file := range files {
		if b, err := cacheContains(link, file); !b || err != nil {
			t.Errorf("cache should contains %s", file)
		}
	}

	invalideCache()

}

func TestCacheErrorCases(t *testing.T) {
	err := cacheAdd("", "file1")
	if err == nil {
		t.Error("invalid action should return an error")
	}

	err = cacheAdd(copy, "")
	if err == nil {
		t.Error("invalid file pathshould return an error")
	}

}
