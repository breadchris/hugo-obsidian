package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gernest/front"
)

func findUniqueShortnames(shortname, source, existingPath string) (newSourceShortname, newExistingShortname string) {
	/*
		/animals/cats/paws
		/animals/dogs/paws

		shortname collision: paws
		updated shortnames: cats/paws, dogs/paws
	*/

	if source == existingPath {
		fmt.Printf("source (%s) and existing path (%s) are the same, not able to create unique shortnames", source, existingPath)
		return
	}

	restOfSource := filepath.Dir(source)
	restOfExisting := filepath.Dir(existingPath)

	commonPath := shortname
	for restOfSource != "/" && restOfExisting != "/" {
		sourceBase := filepath.Base(restOfSource)
		existingBase := filepath.Base(restOfExisting)

		if sourceBase != existingBase {
			newSourceShortname = path.Join(commonPath, sourceBase)
			newExistingShortname = path.Join(commonPath, sourceBase)
			return
		}
		commonPath = path.Join(sourceBase, commonPath)

		restOfSource = filepath.Dir(restOfSource)
		restOfExisting = filepath.Dir(restOfExisting)
	}
	return
}

// recursively walk directory and return all files with given extension
func walk(root, ext string, index bool, ignorePaths map[string]struct{}) (res []Link, i ContentIndex, shortnameToPathLookup map[string]string) {
	fmt.Printf("Scraping %s\n", root)
	i = make(ContentIndex)

	m := front.NewMatter()
	m.Handle("---", front.YAMLHandler)
	nPrivate := 0

	start := time.Now()

	shortnameToPathLookup = map[string]string{}

	err := filepath.WalkDir(root, func(fp string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}

		// path normalize fp
		s := filepath.ToSlash(fp)
		if _, ignored := ignorePaths[s]; ignored {
			fmt.Printf("[Ignored] %s\n", d.Name())
			nPrivate++
		} else if filepath.Ext(d.Name()) == ext {
			res = append(res, parse(s, root)...)
			if index {
				text := getText(s)

				frontmatter, body, err := m.Parse(strings.NewReader(text))
				if err != nil {
					frontmatter = map[string]interface{}{}
					body = text
				}

				var title string
				if parsedTitle, ok := frontmatter["title"]; ok {
					title = parsedTitle.(string)
				} else {
					title = "Untitled Page"
				}

				// check if page is private
				if parsedPrivate, ok := frontmatter["draft"]; !ok || !parsedPrivate.(bool) {
					info, _ := os.Stat(s)
					source := processSource(trim(s, root, ".md"))

					// adjustedPath := UnicodeSanitize(strings.Replace(hugoPathTrim(trim(s, root, ".md")), " ", "-", -1))
					i[source] = Content{
						LastModified: info.ModTime(),
						Title:        title,
						Content:      body,
					}

					shortname := filepath.Base(source)
					if existingPath, ok := shortnameToPathLookup[shortname]; ok {

						// we have a collision with the shortname, lets find unique names for the two
						delete(shortnameToPathLookup, shortname)

						newSourceShortname, newExistingShortname := findUniqueShortnames(shortname, source, existingPath)

						shortnameToPathLookup[newSourceShortname] = source
						shortnameToPathLookup[newExistingShortname] = existingPath
					} else {
						shortnameToPathLookup[shortname] = source
					}
				} else {
					fmt.Printf("[Ignored] %s\n", d.Name())
					nPrivate++
				}
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	end := time.Now()

	fmt.Printf("[DONE] in %s\n", end.Sub(start).Round(time.Millisecond))
	fmt.Printf("Ignored %d private files \n", nPrivate)
	fmt.Printf("Parsed %d total links \n", len(res))
	return
}

func getText(dir string) string {
	// read file
	fileBytes, err := ioutil.ReadFile(dir)
	if err != nil {
		panic(err)
	}

	return string(fileBytes)
}
