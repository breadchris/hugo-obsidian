package main

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"path/filepath"
)

func write(links []Link, contentIndex ContentIndex, shortnameToPathLookup map[string]string, toIndex bool, out string) error {
	index := index(links, shortnameToPathLookup)
	resStruct := struct {
		Index Index  `json:"index"`
		Links []Link `json:"links"`
	}{
		Index: index,
		Links: links,
	}
	marshalledIndex, mErr := json.MarshalIndent(&resStruct, "", "  ")
	if mErr != nil {
		return mErr
	}

	writeErr := ioutil.WriteFile(path.Join(out, "linkIndex.json"), marshalledIndex, 0644)
	if writeErr != nil {
		return writeErr
	}

	if toIndex {
		marshalledContentIndex, mcErr := json.MarshalIndent(&contentIndex, "", "  ")
		if mcErr != nil {
			return mcErr
		}

		writeErr = ioutil.WriteFile(path.Join(out, "contentIndex.json"), marshalledContentIndex, 0644)
		if writeErr != nil {
			return writeErr
		}
	}

	return nil
}

// constructs index from links
func index(links []Link, shortnameToPathLookup map[string]string) (index Index) {
	linkMap := make(map[string][]Link)
	backlinkMap := make(map[string][]Link)
	for _, l := range links {
		// backlink (only if internal)
		if _, ok := backlinkMap[l.Target]; ok {
			backlinkMap[l.Target] = append(backlinkMap[l.Target], l)
		} else {
			// try the shortname for the link
			targetName := filepath.Base(l.Target)

			if path, ok := shortnameToPathLookup[targetName]; ok {
				backlinkMap[path] = append(backlinkMap[path], l)
			} else {
				backlinkMap[l.Target] = []Link{l}
			}
		}

		// regular link
		if _, ok := linkMap[l.Source]; ok {
			linkMap[l.Source] = append(linkMap[l.Source], l)
		} else {
			linkMap[l.Source] = []Link{l}
		}
	}
	index.Links = linkMap
	index.Backlinks = backlinkMap
	return index
}
