package wx

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func mapFromFile(file string) map[string]string {
	m := map[string]string{}
	f, err := os.Open(file)
	if err != nil {
		return m
	}
	err = json.NewDecoder(f).Decode(&m)
	if err != nil {
		fmt.Println("mapToFile error:", err.Error())
	}
	defer f.Close()
	return m
}

func mapToFile(m map[string]string, file string) error {
	f, err := os.Create(file)
	if err != nil {
		log.Printf("Unable to cache oauth token: %v", err)
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(m)
}

type FileStore struct {
	Path string
}

func (fs *FileStore) GetFromFile(key string) string {
	m := mapFromFile(fs.Path)
	if val, ok := m[key]; ok {
		return val
	}
	return ""
}

func (fs *FileStore) SaveToFile(key, value string) error {
	m := mapFromFile(fs.Path)
	m[key] = value
	return mapToFile(m, fs.Path)
}
