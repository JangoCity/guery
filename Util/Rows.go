package Util

import (
	"sort"
)

type Rows struct {
	Data  []*Row
	Order []OrderType
}

func NewRows(order []OrderType) *Rows {
	return &Rows{
		Data:  []*Row{},
		Order: order,
	}
}

func (self *Rows) Min() int {
	res := -1
	ln := len(self.Data)
	for i := 0; i < ln; i++ {
		if self.Data[i] == nil {
			continue
		}
		if res < 0 {
			res = i
		} else {
			if self.Less(i, res) {
				res = i
			}
		}
	}
	return res
}

func (self *Rows) Less(i, j int) bool {
	rowi, rowj := self.Data[i], self.Data[j]
	for k := 0; k < len(self.Order); k++ {
		vi, vj := rowi.Vals[k], rowj.Vals[k]
		if vi == vj {
			continue
		}
		res := Less(vi, vj)
		if self.Order[k] == DESC {
			res = !res
		}
		return res
	}
	return false
}

func (self *Rows) Swap(i, j int) {
	self.Data[i], self.Data[j] = self.Data[j], self.Data[i]
}

func (self *Rows) Len() int {
	return len(self.Data)
}

func (self *Rows) Sort() {
	sort.Sort(self)
}

func (self *Rows) Append(row *Row) {
	self.Data = append(self.Data, row)
}
