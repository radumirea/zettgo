package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var BaseDir string
var NoteDir string
var ImgDir string
var ImgtmpDir string
var DraftDir string
var MetaDir string
var TemplatesDir string
var HtmlDir string
var Editor string
var DraftTemplate string

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	app := getCliConfig()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func finishDraft() error {
	selection, err := listDrafts(true, "Select draft to finish: ")
	if err != nil {
		return err
	}
	if err := compileReferences(DraftDir + selection); err != nil {
		return err
	}
	return os.Rename(DraftDir+selection, NoteDir+selection)
}

func compileReferences(fileName string) error {
	linkRegex := regexp.MustCompile(`\[\[[^][]*\|([^][]*)\]\]`)
	noteBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	note := string(noteBytes)
	noteId := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
	title, _ := getTitleFromDraft(fileName)
	links := linkRegex.FindAllString(note, -1)
	extractRegex := regexp.MustCompile(`\[\[.*\|(.*)\]\]`)
	toMdRegex := regexp.MustCompile(`\[\[([^][]*)\|([^][]*)\]\]`)
	for _, link := range links {
		refId := extractRegex.ReplaceAllString(link, `$1`)
		ref := NoteDir + refId
		exists := checkFileContent(ref, `\[\[[^][]*\|`+noteId+`\]\]|@noBacklink`)
		if !exists {
			if err := appendToFile(ref, "\n- [["+title+"|"+noteId+"]]"); err != nil {
				return err
			}
			refBytes, err := ioutil.ReadFile(ref)
			if err != nil {
				return err
			}
			mdRef := toMdRegex.ReplaceAllString(string(refBytes), `[$1]($2.html)`)
			if err := mdToHtml(mdRef, HtmlDir+refId+".html"); err != nil {
				return err
			}
		}
	}
	note = toMdRegex.ReplaceAllString(note, `[$1]($2.html)`)
	note, err = compileImages(note, noteId)
	if err != nil {
		return err
	}
	return mdToHtml(note, HtmlDir+noteId+".html")
}

func compileImages(note string, id string) (string, error) {
	imgRegex := regexp.MustCompile(`\(\(([^\(\)]*)\)\)`)
	extractRegex := regexp.MustCompile(`\(\((.*)\)\)`)
	imgRefs := imgRegex.FindAllString(note, -1)
	for _, imgRef := range imgRefs {
		imgName := extractRegex.ReplaceAllString(imgRef, `$1`)
		if _, err := os.Stat(ImgDir + id + "-" + imgName); errors.Is(err, os.ErrNotExist) && imgName != "" {
			if err := os.Rename(ImgtmpDir+imgName, ImgDir+id+"-"+imgName); err != nil {
				fmt.Println("Could not move " + imgName + " to image directory")
			}
		} else if err != nil {
			return "", err
		}
	}
	return imgRegex.ReplaceAllStringFunc(note, func(match string) string {
		return imageToBase64Tag(ImgDir + id + "-" + match[2:len(match)-2])
	}), nil
}

func listDrafts(askInput bool, prompt string) (string, error) {
	files, _ := os.ReadDir(DraftDir)
	fmt.Println("Listing drafts:")
	if len(files) == 0 {
		fmt.Println("  No drafts found")
		return "", nil
	}
	errGroup := errors.New("")
	for i := 0; i < len(files); i++ {
		if title, err := getTitleFromDraft(DraftDir + files[i].Name()); err == nil {
			fmt.Println("	" + fmt.Sprint(i+1) + " " + title)
		} else {
			errGroup = errors.New(errGroup.Error() + "\n" + err.Error())
		}
	}
	if errGroup.Error() != "" {
		return "", errGroup
	}
	if askInput {
		if index, err := strconv.Atoi(getUserInput(prompt)); err == nil {
			if index > 0 && index <= len(files) {
				return files[index-1].Name(), nil
			} else {
				return "", errors.New("Selection not in range")
			}
		} else {
			return "", errors.New("Selection needs to be a number")
		}
	} else {
		return "", nil
	}
}

func newDraft() error {
	input, err := ioutil.ReadFile(TemplatesDir + DraftTemplate)
	if err != nil {
		return err
	}
	draftId, err := incMetaCount()
	if err != nil {
		return err
	}
	fileName := DraftDir + fmt.Sprint(draftId)
	err = ioutil.WriteFile(fileName, input, 0644)
	if err != nil {
		return err
	}
	return openEditor(fileName)
}

func editDraft() error {
	selection, err := listDrafts(true, "Select draft to edit: ")
	if err != nil {
		return err
	}
	return openEditor(DraftDir + selection)
}

func openEditor(fileName string) error {
	editorPath, _ := exec.LookPath(Editor)
	cmd := exec.Command(editorPath, fileName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func deleteDraft() error {
	selection, err := listDrafts(true, "Select a draft to delete: ")
	if err != nil {
		return err
	}
	return os.Remove(DraftDir + selection)
}

func deleteNote(id string) error {
	if err := os.Remove(NoteDir + id); err != nil {
		return err
	}
	return os.Remove(HtmlDir + id + ".html")
}

func rewriteNote(id string) error {
	if _, err := os.Stat(NoteDir + id); errors.Is(err, os.ErrNotExist) {
		return errors.New("Note with id " + id + " does not exist")
	} else if err != nil {
		return err
	}
	if err := openEditor(NoteDir + id); err != nil {
		return err
	}
	return compileReferences(NoteDir + id)
}

func recompileAll() error {
	files, _ := os.ReadDir(NoteDir)
	errGroup := errors.New("")
	for i := 0; i < len(files); i++ {
		if !files[i].IsDir() {
			if err := compileReferences(NoteDir + files[i].Name()); err != nil {
				errGroup = errors.New(errGroup.Error() + "\n" + err.Error())
			}
		}
	}
	return nil
}
