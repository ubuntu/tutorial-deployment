// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Some filters imported and renamed from https://github.com/googlecodelabs/tools/blob/master/export.go

package claattools

import "github.com/didrocks/codelab-ubuntu-tools/claat/types"

// GetImageNodes filters out everything except types.NodeImage nodes, recursively.
func GetImageNodes(nodes []types.Node) []*types.ImageNode {
	var imgs []*types.ImageNode
	for _, n := range nodes {
		switch n := n.(type) {
		case *types.ImageNode:
			imgs = append(imgs, n)
		case *types.ListNode:
			imgs = append(imgs, GetImageNodes(n.Nodes)...)
		case *types.ItemsListNode:
			for _, i := range n.Items {
				imgs = append(imgs, GetImageNodes(i.Nodes)...)
			}
		case *types.HeaderNode:
			imgs = append(imgs, GetImageNodes(n.Content.Nodes)...)
		case *types.URLNode:
			imgs = append(imgs, GetImageNodes(n.Content.Nodes)...)
		case *types.ButtonNode:
			imgs = append(imgs, GetImageNodes(n.Content.Nodes)...)
		case *types.InfoboxNode:
			imgs = append(imgs, GetImageNodes(n.Content.Nodes)...)
		case *types.GridNode:
			for _, r := range n.Rows {
				for _, c := range r {
					imgs = append(imgs, GetImageNodes(c.Content.Nodes)...)
				}
			}
		}
	}
	return imgs
}

// GetImportNodes filters out everything except types.NodeImport nodes, recursively.
func GetImportNodes(nodes []types.Node) []*types.ImportNode {
	var imps []*types.ImportNode
	for _, n := range nodes {
		switch n := n.(type) {
		case *types.ImportNode:
			imps = append(imps, n)
		case *types.ListNode:
			imps = append(imps, GetImportNodes(n.Nodes)...)
		case *types.InfoboxNode:
			imps = append(imps, GetImportNodes(n.Content.Nodes)...)
		case *types.GridNode:
			for _, r := range n.Rows {
				for _, c := range r {
					imps = append(imps, GetImportNodes(c.Content.Nodes)...)
				}
			}
		}
	}
	return imps
}
