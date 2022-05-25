package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func formatPath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	path = filepath.FromSlash(path)
	if path == "~" {
		return dir
	} else if strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		path = filepath.Join(dir, path[2:])
	}
	if path[len(path)-1] != os.PathSeparator {
		path += string(os.PathSeparator)
	}
	return path
}

func getUserInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func getTitleFromDraft(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		match, _ := regexp.MatchString("^[[:space:]]*#[[:space:]].+", scanner.Text())
		if match {
			return strings.Split(scanner.Text(), "# ")[1], nil
		}
	}
	return "Untitled (" + fmt.Sprint(filepath.Base(fileName)) + ")", nil
}

func checkFileContent(fileName, content string) bool {
	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return true
	}
	file := string(fileBytes)
	exists, _ := regexp.MatchString(file, content)
	return exists
}

func appendToFile(fileName, content string) error {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = file.WriteString(content); err != nil {
		return err	
	}	
	return nil
}

func incMetaCount() (int, error) {
	input, err := ioutil.ReadFile(MetaDir + "count")
	if err != nil {
		return -1, errors.New("Corrupted metadata, please regenerate")
	}
	count, _ := strconv.Atoi(string(input))
	count++
	err = ioutil.WriteFile(MetaDir+"count", []byte(fmt.Sprint(count)), 0644)
	if err != nil {
		return -1, errors.New("Error writing to " + MetaDir + "count")
	}
	return count, nil
}

func deleteFile(file string) error {
	return nil
}

func mdToHtml(source, dest string) error {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(dest, buf.Bytes(), 0644)
}
