package watcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Directory struct {
	Path string `json:"path,omitempty"`
	Hash uint32 `json:"hash,omitempty"`
}

func (w *Watcher) loadDirs() {
	jsonFile, err := os.Open(dirsFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(byteValue, &w.Dirs)
	if err != nil {
		fmt.Println(err)
	}
}

func (w *Watcher) dumpDirs() {
	rankingsJson, _ := json.Marshal(w.Dirs)
	_ = ioutil.WriteFile(dirsFile, rankingsJson, 0o644)
}
