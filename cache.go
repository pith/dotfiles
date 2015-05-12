package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Cache represents the structure of cache data
type Cache struct {
	Link         []string
	Copy         []string
	InitSelected []string
	InitRun      []string
}

// Action is a type of action that can be cached
type Action string

// List of the actions cached
const (
	link         Action = "link"
	copy         Action = "copy"
	initSelected Action = "initSelected"
	initRun      Action = "initRun"
)

var (
	cachePath = filepath.Join(BaseDir, "cache", "cache.json")
	cache     Cache
)

func loadCache() {
	if _, err := os.Stat(filepath.Join(BaseDir, "cache")); os.IsNotExist(err) {
		err := os.Mkdir(filepath.Join(BaseDir, "cache"), 0777)
		if err != nil {
			log.Fatal("Failed to create cache dir: ", err)
		}
	}

	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		// Cache exists, load it
		bytes, err := ioutil.ReadFile(cachePath)
		if err != nil {
			log.Fatalf("Unable to load the cache:\n%s", err)
		}

		err = json.Unmarshal(bytes, &cache)
		if err != nil {
			log.Fatalf("Failed to unmarshall the cache file:\n%s", err)
		}

	}
}

func cacheAdd(action Action, file string) error {
	if file == "" {
		return fmt.Errorf("The path to the cache file cannot be \"\"")
	}

	switch action {
	case link:
		cache.Link = append(cache.Link, file)
	case copy:
		cache.Copy = append(cache.Copy, file)
	case initSelected:
		cache.InitSelected = append(cache.InitSelected, file)
	case initRun:
		cache.InitRun = append(cache.InitRun, file)
	default:
		return fmt.Errorf("%s is not part of the possible cached actions", action)
	}

	flushCache()

	return nil
}

func cacheContains(action Action, file string) (bool, error) {
	if file == "" {
		return false, fmt.Errorf("The path to the cache file cannot be \"\"")
	}

	var res bool
	switch action {
	case link:
		res = stringSlice(cache.Link).indexOf(file) != -1
	case copy:
		res = stringSlice(cache.Copy).indexOf(file) != -1
	case initSelected:
		res = stringSlice(cache.InitSelected).indexOf(file) != -1
	case initRun:
		res = stringSlice(cache.InitRun).indexOf(file) != -1
	default:
		return false, fmt.Errorf("%s is not part of the possible cached actions", action)
	}

	return res, nil
}

func cacheRemove(action Action, file string) error {
	if file == "" {
		return fmt.Errorf("The path to the cache file cannot be \"\"")
	}

	switch action {
	case link:
		cache.Link = stringSlice(cache.Link).remove(file)
	case copy:
		cache.Copy = stringSlice(cache.Copy).remove(file)
	case initSelected:
		cache.InitSelected = stringSlice(cache.InitSelected).remove(file)
	case initRun:
		cache.InitRun = stringSlice(cache.InitRun).remove(file)
	default:
		return fmt.Errorf("%s is not part of the possible cached actions", action)
	}

	flushCache()

	return nil
}

// Write the cache on disk
func flushCache() {
	bytes, err := json.Marshal(&cache)
	if err != nil {
		log.Fatal("Unable to marshal the cache: ", err)
	}

	if _, err := os.Stat(filepath.Join(BaseDir, "cache")); os.IsNotExist(err) {
		loadCache()
	}

	err = ioutil.WriteFile(cachePath, bytes, 0666)
	if err != nil {
		log.Fatal("Unable to write the cache: ", err)
	}
}

func invalideCache() error {
	cache = Cache{}
	return os.Remove(cachePath)
}

type stringSlice []string

func (slice stringSlice) indexOf(elem string) int {
	for i, item := range slice {
		if item == elem {
			return i
		}
	}
	return -1
}

func (slice stringSlice) remove(elem string) stringSlice {
	index := slice.indexOf(elem)
	return append(slice[:index], slice[index+1:]...)
}
