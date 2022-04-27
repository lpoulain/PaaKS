package main

import (
	"io"
	"os"
)

func createServiceFiles(name string) error {
	err := os.Mkdir("/tmp/storage/"+name, 0755)

	src := "/tmp/storage/template/handler.py"
	dst := "/tmp/storage/" + name + "/handler.py"
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return err
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)

	if err != nil {
		return err
	}

	return nil
}

func deleteServiceFiles(name string) error {
	return os.RemoveAll("/tmp/storage/" + name)
}
