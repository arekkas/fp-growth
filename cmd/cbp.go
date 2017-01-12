package cmd

import "sort"

type ConditionalPatternBases []ConditionalPatternBase

type ConditionalPatternBase struct {
	Prefix ConditionalItem
	Bases  [][]ConditionalItem
}

type ConditionalItem struct {
	Item  int
	Count int
}

func mineParents(p *FPTreeNode, max int) (pb []ConditionalItem) {
	for p != nil {
		count := p.Count
		if count > max {
			count = max
		}

		pb = append(pb, ConditionalItem{Item: p.Item, Count: count})
		p = p.Parent
	}
	return pb
}

func MineConditionalPatternBases(t FPTree, ht HeadTable) ConditionalPatternBases {
	base := ConditionalPatternBases{}

	for i := len(ht) - 1; i >= 0; i-- {
		l := ht[i].Link
		branch := 0
		pb := ConditionalPatternBase{
			Prefix: ConditionalItem{
				Item: l.Item,
				Count: ht[i].Count,
			},
			Bases: [][]ConditionalItem{{}},
		}

		if l.Parent != nil {
			pb.Bases[branch] = mineParents(l.Parent, l.Count)
		}

		l = l.Link
		for l != nil {
			count := l.Count
			if count > pb.Prefix.Count {
				count = pb.Prefix.Count
			}

			if l.Parent != nil {
				pb.Bases = append(pb.Bases, []ConditionalItem{})
				branch++
				pb.Bases[branch] = mineParents(l.Parent, l.Count)
			}
			l = l.Link
		}

		base = append(base, pb)
	}

	return base
}

type ConditionalFPTree struct {
	Prefix ConditionalItem
	Tree FPTree
}

func ConstructConditionalFPTrees(bs ConditionalPatternBases, ht ConditionalHeadTables) (res []ConditionalFPTree) {
	for _, tx := range bs {
		links := map[int][]*FPTreeNode{}
		t := ConditionalFPTree {
			Prefix: tx.Prefix,
			Tree: FPTree{
				Root: &FPTreeNode{
					Children: []*FPTreeNode{},
				},
			},
		}

		var items Items
		for _, b := range tx.Bases {
			items = make(Items, len(b))
			for k, v := range b {
				items[k] = v.Item
			}

			buildTree(items, t.Tree.Root, nil, links)
		}

		for item, l := range links {
			ht.Get(tx.Prefix.Item).SetLink(item, l[0])
			if len(l) == 1 {
				continue
			}

			for k, n := range l[0:len(l) - 1] {
				n.Link = l[k+1]
			}
		}

		res = append(res, t)
	}

	return res
}

type OrderableItems []ConditionalItem

func (p OrderableItems) Len() int {
	return len(p)
}
func (p OrderableItems) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type OrderableItemsWrapper struct {
	OrderableItems
	HT ConditionalHeadTable
}

func (p OrderableItemsWrapper) Less(i, j int) bool {
	if p.OrderableItems[i].Count == p.OrderableItems[i].Count {
		if p.HT.GetPosition(p.OrderableItems[i].Item) == p.HT.GetPosition(p.OrderableItems[j].Item) {
			return p.OrderableItems[i].Item > p.OrderableItems[j].Item
		}
		return p.HT.GetPosition(p.OrderableItems[i].Item) < p.HT.GetPosition(p.OrderableItems[j].Item)
	}
	return p.OrderableItems[i].Count < p.OrderableItems[j].Count
}

type ConditionalHeadTables []ConditionalHeadTable

func (ht ConditionalHeadTables) GetIndex(item int) int {
	for i, h := range ht {
		if h.Prefix.Item == item {
			return i
		}
	}

	return -1
}
func (ht ConditionalHeadTables) Get(item int) *ConditionalHeadTable {
	return &ht[ht.GetIndex(item)]
}

func OrderConditionalPatternBases(bs ConditionalPatternBases, ht ConditionalHeadTables) ConditionalPatternBases {

	// https://godoc.org/sort#example-package--SortWrapper

	nbs := make(ConditionalPatternBases, len(bs))
	for x, b := range bs {
		htindex := ht.GetIndex(b.Prefix.Item)

		nbs[x].Prefix = ht[htindex].Prefix
		for _, bb := range b.Bases {
			w := OrderableItemsWrapper{OrderableItems: OrderableItems(bb), HT: ht[htindex]}
			sort.Sort(w)
			for i, o := range w.OrderableItems {
				if ht[htindex].Get(o.Item) == nil {
					w.OrderableItems=append(w.OrderableItems[:i], w.OrderableItems[i+1:]...)
				}
			}
			nbs[x].Bases = append(nbs[x].Bases, []ConditionalItem(w.OrderableItems))
		}

	}
	return nbs
}

type ConditionalHeadTable struct {
	Prefix ConditionalItem
	HeadTable
}

func ConstructConditionalHeadTables(bs ConditionalPatternBases, minSup int) []ConditionalHeadTable {
	var tables []ConditionalHeadTable
	for _, base := range bs {
		ic := map[int]int{}
		for _, iss := range base.Bases {

			for _, tx := range iss {
				ic[tx.Item] = ic[tx.Item] + tx.Count
			}
		}

		for k, c := range ic {
			if c < minSup {
				delete(ic, k)
			}
		}

		pl := make(HeadTable, len(ic))
		i := 0
		for k, v := range ic {
			pl[i] = HeadTableRow{Item: k, Count: v}
			i++
		}
		sort.Sort(sort.Reverse(pl))
		tables = append(tables, ConditionalHeadTable{
			Prefix: base.Prefix,
			HeadTable: pl,
		})
	}
	return tables
}