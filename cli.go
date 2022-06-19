package main

import (
	"github.com/urfave/cli/v2"
)

func getCliConfig() *cli.App{
	return &cli.App{
		Name: "zettgo",
		Before: func(ctx *cli.Context) error {
			BaseDir	= formatPath(BaseDir)
			TemplatesDir = formatPath(BaseDir + TemplatesDir)
			NoteDir = formatPath(BaseDir + NoteDir)
			ImgDir = formatPath(BaseDir + ImgDir)
			ImgtmpDir = formatPath(BaseDir + ImgtmpDir)
			DraftDir = formatPath(BaseDir + DraftDir)
			MetaDir = formatPath(BaseDir + MetaDir)
			HtmlDir = formatPath(BaseDir + HtmlDir)
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "basedir",
				Value:       "~/.zettgo/",
				Usage:       "base directory",
				Destination: &BaseDir,
			},
			&cli.StringFlag{
				Name:        "editor",
				Value:       "vim",
				Usage:       "the text editor",
				EnvVars:     []string{"EDITOR"},
				Destination: &Editor,
			},
			&cli.StringFlag{
				Name:        "notedir",
				Value:       "notes",
				Usage:       "directory for notes",
				Destination: &NoteDir,
			},
			&cli.StringFlag{
				Name:        "imgdir",
				Value:       "notes/imgs",
				Usage:       "directory for images",
				Destination: &ImgDir,
			},
			&cli.StringFlag{
				Name:        "draftdir",
				Value:       "drafts",
				Usage:       "directory for storing drafts",
				Destination: &DraftDir,
			},
			&cli.StringFlag{
				Name:        "templatedir",
				Value:       "templates",
				Usage:       "directory for storing templates",
				Destination: &TemplatesDir,
			},
			&cli.StringFlag{
				Name:        "configdir",
				Value:       "config",
				Usage:       "directory for config and metadata files",
				Destination: &MetaDir,
			},
			&cli.StringFlag{
				Name:        "imgtmp",
				Value:       "imgtmp",
				Usage:       "location for fetching images on draft finish",
				Destination: &ImgtmpDir,
			},
			&cli.StringFlag{
				Name:        "htmldir",
				Value:       "html",
				Usage:       "location for storing html compiled notes",
				Destination: &HtmlDir,
			},
			&cli.StringFlag{
				Name:        "drafttemplate",
				Value:       "draftTemplate.md",
				Usage:       "template to be used when creating a new draft",
				Destination: &DraftTemplate,
			},
		}, Usage: "zettelkasten note taking tool",
		Commands: []*cli.Command{
			{
				Name:    "n",
				Aliases: []string{"new"},
				Usage:   "start a new draft",
				Action: func(c *cli.Context) error {
					return newDraft()
				},
			},
			{
				Name:    "e",
				Aliases: []string{"edit"},
				Usage:   "edit draft",
				Action: func(c *cli.Context) error {
					return editDraft()
				},
			},
			{
				Name:    "dd",
				Aliases: []string{"deld"},
				Usage:   "delete draft",
				Action: func(c *cli.Context) error {
					return deleteDraft()
				},
			},
			{
				Name:    "l",
				Aliases: []string{"list"},
				Usage:   "list drafts",
				Action: func(c *cli.Context) error {
					_, err := listDrafts(false, "")
					return err
				},
			},
			{
				Name:    "dn",
				Aliases: []string{"deln"},
				Usage:   "delete note",
				ArgsUsage: "[note_id]",
				Action: func(c *cli.Context) error {
					if c.NArg() > 0 {
						return deleteNote(c.Args().First())
					} else {
						selection := getUserInput("Input a note id: ")
						return deleteNote(selection)
					}
				},
			},
			{
				Name:    "f",
				Aliases: []string{"finish"},
				Usage:   "finish draft and convert to note",
				Action: func(c *cli.Context) error {
					return finishDraft()
				},
			},
			{
				Name:    "r",
				Aliases: []string{"rewrite"},
				Usage:   "rewrite note",
				Action: func(c *cli.Context) error {
					if c.NArg() > 0 {
						return rewriteNote(c.Args().First())
					} else {
						selection := getUserInput("Input a note id: ")
						return rewriteNote(selection)
					}
				},
			},
			{
				Name:    "recompile",
				Usage:   "recompile all notes",
				Action: func(c *cli.Context) error {
					return recompileAll()
				},
			},
		},
	}
}