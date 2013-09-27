/*
  Copyright 2012 Sergey Cherepanov (https://github.com/cheggaaa)

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package storage

import (
	"errors"
	"fmt"
)

var ErrFileExists = errors.New("File exists")
var ErrFileNotFound = errors.New("File not found")

const WILDCARD = "*"

type Node struct {
	File *File
	Childs map[string]*Node
}

func (n *Node) IsFile() bool {
	return n.File != nil
}

func (n *Node) Get(parts []string, depth int) (f *File, err error) {
	// is it target
	if len(parts) == depth {
		if ! n.IsFile() {
			err = ErrFileNotFound
			return
		}
		f = n.File
		return
	}
	// find in childs
	if n.Childs != nil {
		if child, ok := n.Childs[parts[depth]]; ok {
			return child.Get(parts, depth + 1)
		} 
	}
	err = ErrFileNotFound
	return
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
	if n.Childs == nil {
		err = ErrFileNotFound
		return 
	}
	
	// is last - delete
	if len(parts) - 1 == depth {
		if child, ok := n.Childs[parts[depth]]; ok {
			if child.IsFile() {
				f = child.File
				child.File = nil
				if child.Childs == nil || len(child.Childs) == 0 {
					delete(n.Childs, parts[depth])
				}
				return
			}
		} 
		err = ErrFileNotFound
		return
	}

	// to child
	if child, ok := n.Childs[parts[depth]]; ok {
		if f, err = child.Delete(parts, depth + 1); err == nil {
			if ! child.IsFile() && (child.Childs == nil || len(child.Childs) == 0) {
				delete(n.Childs, parts[depth])
			}
		} else {
			return
		}
	} else {
		err = ErrFileNotFound
	}
	return
}

func (n *Node) List(parts []string, depth, nesting int) (files []string, err error) {
	files = make([]string, 0)	
	// make list from childs
	if len(parts) <= depth {
		if n.Childs != nil {
			for name, node := range n.Childs {
				if node.IsFile() {
					files = append(files, name)
				}
				childFiles, _ := node.List(parts, depth, nesting)
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
			if files, err = node.List(parts, depth + 1, nesting); err == nil {
				for i, name := range files {
					files[i] = parts[depth] + "/" + name
				}
			} else {
				return
			}
		} else {
			err = ErrFileNotFound
		}
	} else {
		err = ErrFileNotFound
	}
	return
} 


func (n *Node) Print(prefix string) {
	p := "E"
	if n.IsFile() {
		p = "F"
	}
	fmt.Printf("(%s) %s\n", p, prefix)
	if n.Childs != nil {
		for name, child := range n.Childs {
			child.Print(prefix + "/" + name)
		}
	}
}