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

	"github.com/urfave/cli/v2"
)

var BaseDir string
var NoteDir string
var ImgDir string
var ImgtmpDir string
var DraftDir string
var MetaDir string
var TemplatesDir string
var Editor string

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	app := &cli.App{
		Name: "zettgo",
		Before: func(ctx *cli.Context) error {
			TemplatesDir = formatPath(TemplatesDir)
			NoteDir = formatPath(NoteDir)
			ImgDir = formatPath(ImgDir)
			ImgtmpDir = formatPath(ImgtmpDir)
			DraftDir = formatPath(DraftDir)
			MetaDir = formatPath(MetaDir)
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "editor",
				Value:       "vim",
				Usage:       "the text editor",
				EnvVars:     []string{"EDITOR"},
				Destination: &Editor,
			},
			&cli.StringFlag{
				Name:        "notedir",
				Value:       "~/.zettgo/notes",
				Usage:       "directory for notes",
				Destination: &NoteDir,
			},
			&cli.StringFlag{
				Name:        "imgdir",
				Value:       "~/.zettgo/notes/imgs",
				Usage:       "directory for images",
				Destination: &ImgDir,
			},
			&cli.StringFlag{
				Name:        "draftdir",
				Value:       "~/.zettgo/drafts",
				Usage:       "directory for storing drafts",
				Destination: &DraftDir,
			},
			&cli.StringFlag{
				Name:        "templatedir",
				Value:       "~/.zettgo/templates",
				Usage:       "directory for storing templates",
				Destination: &TemplatesDir,
			},
			&cli.StringFlag{
				Name:        "configdir",
				Value:       "~/.zettgo/config",
				Usage:       "directory for config and metadata files",
				Destination: &MetaDir,
			},
			&cli.StringFlag{
				Name:        "imgtmp",
				Value:       "~/.zettgo/imgtmp",
				Usage:       "location for fetching images on draft finish",
				Destination: &ImgtmpDir,
			},
		}, Usage: "zettelkasten note taking tool",
		Commands: []*cli.Command{
			{
				Name:    "new",
				Aliases: []string{"n"},
				Usage:   "start a new draft",
				Action: func(c *cli.Context) error {
					err := newDraft()
					return err
				},
			},
			{
				Name:    "edit",
				Aliases: []string{"w"},
				Usage:   "edit draft",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "delete draft",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:    "delete_note",
				Aliases: []string{"dd"},
				Usage:   "delete note",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:    "finish",
				Aliases: []string{"f"},
				Usage:   "finish draft and convert to note",
				Action: func(c *cli.Context) error {
					return finishDraft()
				},
			},
			{
				Name:    "rewrite",
				Aliases: []string{"r"},
				Usage:   "rewrite note",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func finishDraft() error {
	selection, err := listDrafts(true, "Select a draft to finish: ")
	if err != nil {
		return err
	}
	if err := computeReferences(DraftDir + selection); err != nil {
		return err
	}
	return os.Rename(DraftDir + selection, NoteDir + selection)
}

func computeReferences(fileName string) error {
	linkRegex := regexp.MustCompile(`\[\[[^][]*\|([^][]*)\]\]`)
	noteBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	note := string(noteBytes)
	noteId := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
	title, _ := getTitleFromDraft(fileName)
	links := linkRegex.FindStringSubmatch(note)
	extractRegex := regexp.MustCompile(`\[\[.*\|(.*)\]\]`)
	toMdRegex := regexp.MustCompile(`\[\[([^][]*)\|([^][]*)\]\]`)
	for _, link := range links {
		refId := extractRegex.ReplaceAllString(link, `$1`)
		ref := NoteDir + string(os.PathSeparator) + refId
		exists := checkFileContent(ref, `\[\[[^][]*\|`+refId+`\]\]`)
		if !exists {
			if err := appendToFile(fileName, `\n- [[`+title+`|`+noteId+`]]`); err != nil {
				return err
			}
			refBytes, err := ioutil.ReadFile(fileName)
			if err != nil {
				return err
			}
			mdRef := toMdRegex.ReplaceAllString(string(refBytes), `\[$1\]($2.html)`)
			if err := mdToHtml(mdRef, NoteDir+refId+".html"); err != nil {
				return err
			}
		}
	}
	mdNote := toMdRegex.ReplaceAllString(note, `\[$1\]($2.html)`)
	return mdToHtml(mdNote, NoteDir+noteId+".html")
}

func listDrafts(askInput bool, prompt string) (string, error) {
	files, _ := ioutil.ReadDir(DraftDir)
	fmt.Println("Listing drafts:")
	if len(files) == 0 {
		fmt.Println("  No drafts found")
		return "", nil
	}
	if len(files) == 1 {
		if title, err := getTitleFromDraft(DraftDir + files[0].Name()); err == nil {
			fmt.Println("  Only one draft found, defaulting to " + title)
			return files[0].Name(), nil
		} else {
			return "", err
		}
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
	input, err := ioutil.ReadFile(TemplatesDir + "draftTemplate.md")
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
	openEditor(fileName)
	return nil
}

func openEditor(fileName string) {
	editorPath, _ := exec.LookPath(Editor)
	cmd := exec.Command(editorPath, fileName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func deleteDraft() error {
	selection, err := listDrafts(true, "Select a draft to delete: ")
	if err != nil {
		return err
	}
	deleteFile(DraftDir + selection)
	return nil
}
