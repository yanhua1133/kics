package json

import (
	"strconv"
	"strings"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/relex/aini"
)

// Parser defines a parser type
type Parser struct {
}

func (p *Parser) Resolve(fileContent []byte, filename string) ([]byte, error) {
	return fileContent, nil
}

// Parse parses .ini file and returns it as a Document
func (p *Parser) Parse(filePath string, fileContent []byte) ([]model.Document, []int, error) {
	model.NewIgnore.Reset()

	inventoryReader := strings.NewReader(string(fileContent))
	var inventory, err = aini.Parse(inventoryReader)
	if err != nil {
		return nil, nil, err
	}

	doc := model.Document{}

	childrenMap, _ := refactorInv(inventory.Groups, 1)

	allMap := emptyDocument()
	(*allMap)["children"] = childrenMap
	doc["all"] = allMap

	return []model.Document{doc}, []int{}, nil
}

// refactorInv removes all extra information
func refactorInv(groups map[string]*aini.Group, parentSize int) (doc *model.Document, children map[string]bool) {
	doc = emptyDocument()
	children = make(map[string]bool)
	for _, group := range groups {
		if parentSize != len(group.Parents) {
			continue
		}
		groupMap := emptyDocument()

		ans, childGroup := refactorInv(group.Children, parentSize+1)
		if len(*ans) > 0 {
			(*groupMap)["children"] = ans
		}

		ans = refactorHosts(group.Hosts, childGroup)
		if len(*ans) > 0 {
			(*groupMap)["hosts"] = ans
		}

		children[group.Name] = true
		for child := range childGroup {
			children[child] = true
		}

		(*doc)[group.Name] = groupMap
	}
	return doc, children
}

// refactorHosts only add Hosts that aren't defined in Children
func refactorHosts(hosts map[string]*aini.Host, children map[string]bool) *model.Document {
	hostMap := emptyDocument()
	for _, host := range hosts {
		if !children[host.Name] {
			(*hostMap)[host.Name] = refactorVars(host.Vars)
			children[host.Name] = true
		}
	}
	return hostMap
}

// refactorVars try to convert to float and add all vars
func refactorVars(vars map[string]string) *model.Document {
	varMap := emptyDocument()
	for key, value := range vars {
		valueInt, ok := strconv.ParseFloat(value, 64)

		if ok == nil {
			(*varMap)[key] = valueInt
		} else {
			(*varMap)[key] = value
		}
	}
	return varMap
}

// SupportedExtensions returns extensions supported by this parser, which are only ini extension
func (p *Parser) SupportedExtensions() []string {
	return []string{".ini"}
}

// SupportedTypes returns types supported by this parser, which are ansible - hosts
func (p *Parser) SupportedTypes() map[string]bool {
	return map[string]bool{
		"ansible - hosts": true,
	}
}

// GetKind returns YAML constant kind
func (p *Parser) GetKind() model.FileKind {
	return model.KindINI
}

// GetCommentToken return the comment token of YAML - #
func (p *Parser) GetCommentToken() string {
	return "#"
}

// GetResolvedFiles returns resolved files
func (p *Parser) GetResolvedFiles() map[string]model.ResolvedFile {
	return make(map[string]model.ResolvedFile)
}

// StringifyContent converts original content into string formatted version
func (p *Parser) StringifyContent(content []byte) (string, error) {
	return string(content), nil
}

func emptyDocument() *model.Document {
	return &model.Document{}
}
