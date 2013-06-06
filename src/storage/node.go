package storage

import (
	"errors"
)

var ErrFileExists = errors.New("File exists")
var ErrFileNotFound = errors.New("File not found")

type Node struct {
	File *File
	Childs map[string]*Node
}

func (n *Node) IsFile() bool {
	return n.File != nil
}

func (n *Node) Add(parts []string, f *File, depth int) (err error) {
	// is it last part - add file
	if len(parts) == depth {
		if n.IsFile() {
			return ErrFileExists
		}
		n.File = f
		return
	}
	
	// add/create to child
	if n.Childs == nil {
		n.Childs = make(map[string]*Node)
	}
	
	node, ok := n.Childs[parts[depth]]
	if ! ok {
		node = &Node{}
		n.Childs[parts[depth]] = node
	}
	
	err = node.Add(parts, f, depth + 1)	
	return
}

func (n *Node) Delete(parts []string, depth int) (f *File, err error) {
	// is last - delete
	if len(parts) - 1 == depth {
		if child, ok := n.Childs[parts[depth]]; ok {
			if child.IsFile() {
				f = child.File
				delete(n.Childs, parts[depth])
				return
			}
		} else {
			err = ErrFileNotFound
			return 
		}
	}
	
	// to child
	if child, ok := n.Childs[parts[depth]]; ok {
		return child.Delete(parts, depth + 1)
	} else {
		err = ErrFileNotFound
	}
	return
}

func (n *Node) List(parts []string, depth int) (files []string, err error) {
	files = make([]string, 0)
	// make list from childs
	if len(parts) <= depth {
		if n.Childs != nil {
			for name, node := range n.Childs {
				if node.IsFile() {
					files = append(files, name)
				}
				childFiles, _ := node.List(parts, depth)
				for _, childName := range childFiles {
					files = append(files, name + "/" + childName)
				}
			}
		}
		return
	}
	// or find and call child
	if n.Childs != nil {
		if node, ok := n.Childs[parts[depth]]; ok {
			return node.List(parts, depth + 1)
		} else {
			err = ErrFileNotFound
		}
	} else {
		err = ErrFileNotFound
	}
	return
} 
