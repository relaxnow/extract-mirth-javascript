package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/antchfx/xmlquery"
)

func main() {
	// Figure out where we need to find the XML files and dump out the JS
	searchDir := "."
	if len(os.Args) > 1 {
		searchDir = os.Args[1]
	} else {
		executableDir, err := os.Executable()

		if err != nil {
			panic(err)
		}

		searchDir = executableDir
	}

	searchDirInfo, err := os.Stat(searchDir)

	if err != nil {
		log.Fatal(err)
	}

	if !searchDirInfo.IsDir() {
		log.Fatal("Not a directory", searchDir)
	}

	extractedDir := searchDir + "/extracted-mirth-javascript"
	os.RemoveAll(extractedDir)
	err = os.Mkdir(extractedDir, 0700)

	if err != nil {
		log.Fatal(err)
	}

	fileList := make([]string, 0)
	e := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		if matched, err := filepath.Match("*.xml", filepath.Base(path)); err != nil {
			return err
		} else if matched {
			fileList = append(fileList, path)
		}
		return nil
	})

	if e != nil {
		panic(e)
	}

	for _, filePath := range fileList {
		fmt.Printf("FILE: '%s'\n", filePath)
		file, err := os.Open(filePath)

		if err != nil {
			log.Fatal(err)
		}

		xmlDoc, err := xmlquery.Parse(file)

		if err != nil {
			log.Fatal(err)
		}

		channelNodes := xmlquery.Find(xmlDoc, "/channel")

		for _, channelNode := range channelNodes {
			channelName := channelNode.SelectElement("name").InnerText()
			fmt.Printf("CHANNEL: '%s'\n", channelName)

			channelId := channelNode.SelectElement("id").InnerText()
			fmt.Printf("CHANNEL ID: '%s'\n", channelId)

			channelDir := extractedDir + "/" + channelName
			os.Mkdir(channelDir, 0700)

			codeTemplatesDir := channelDir + "/codeTemplates"
			os.Mkdir(codeTemplatesDir, 0700)

			codeTemplateNodes := xmlquery.Find(channelNode, "exportData/codeTemplateLibraries/codeTemplateLibrary/codeTemplates/codeTemplate")

			for _, codeTemplateNode := range codeTemplateNodes {
				codeTemplateName := codeTemplateNode.SelectElement("name").InnerText()
				fmt.Printf("CodeTemplate named '%s'\n", codeTemplateName)

				codeTemplateId := codeTemplateNode.SelectElement("id").InnerText()
				fmt.Printf("CodeTemplate id '%s'\n", codeTemplateId)

				codeNode := codeTemplateNode.SelectElement("properties/code")

				if codeNode == nil {
					continue
				}

				code := codeNode.InnerText()
				fmt.Printf("Code: '%s'\n", code)

				codeTemplateDir := codeTemplatesDir + "/" + codeTemplateName
				os.Mkdir(codeTemplateDir, 0700)

				jsFileName := codeTemplateDir + "/" + codeTemplateName + ".js"
				fmt.Printf("WRITING FILE '%s'\n", jsFileName)
				os.WriteFile(jsFileName, []byte(code), fs.FileMode(0600))
			}
		}
	}
}
