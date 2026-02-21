package code

import (
	"fmt"
	"io"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
)

type Node struct {
	Parent   *Node
	Body     string
	Children []*Node
}

func NewNode(parent *Node, body string) *Node {
	return &Node{
		Parent:   parent,
		Body:     body,
		Children: []*Node{},
	}
}

func (node *Node) IsRoot() bool {
	return node.Parent == nil
}

func (node *Node) Append(child *Node) {
	node.Children = append(node.Children, child)
}

func (node *Node) ExtractResource() string {
	fields := strings.Fields(node.Body)
	var previous *string
	for i := len(fields) - 1; i >= 0; i-- {
		if fields[i] == "function" {
			if previous == nil {
				return ""
			}
			return *previous
		}
		previous = &fields[i]
	}
	return ""
}

func (node *Node) Format(location string) *Node {
	node.Body = strings.ReplaceAll(node.Body, "./", location+"/")
	return node
}

func (node *Node) EndsWithBackslash() bool {
	line := strings.TrimRight(node.Body, " ")
	return strings.HasSuffix(line, "\\")
}

func (node *Node) Print(writer io.Writer, indent int) {
	WriteIndentString(writer, indent)
	fmt.Fprintln(writer, node.Body)

	for _, child := range node.Children {
		child.Print(writer, indent+4)
	}
}

func (node *Node) String() string {
	var builder strings.Builder
	node.Print(&builder, 0)
	return builder.String()
}

func (node *Node) getRoot() *Node {
	if node.IsRoot() {
		return node
	}
	return node.Parent.getRoot()
}

func (node *Node) MakeErrorContext(i int) liberrors.Context {
	var buffer string
	var buffer_offset int
	if node.Parent != nil && !node.Parent.IsRoot() {
		buffer = node.Parent.Body + "\n"
		buffer_offset++
	}
	if i != 1 {
		buffer += "...\n"
		buffer_offset++
	}
	return liberrors.FileContext{
		Trace: []liberrors.TraceItem{
			{
				Name: node.getRoot().Body,
				Col:  -1,
				Row:  i,
			},
		},
		Buffer: liberrors.Buffer{
			FirstLine:   uint(max(1, i-buffer_offset)),
			Buffer:      buffer,
			Highlighted: node.Body,
		},
	}
}
