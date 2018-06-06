package util

import (
	"testing"
	"fmt"
)

type BlockData struct {
	Height int64 `json:"height"`
	Name   string
}

func TestSend(t *testing.T) {

	AlertMail("etc_payout", "send err", "nadfma")
}

func TestAdd(t *testing.T) {
	a := []*BlockData{}
	b := []*BlockData{}
	c := []*BlockData{}

	s := [][]*BlockData{}
	for i := 0; i < 10; i++ {
		d := &BlockData{Height: int64(i + 3*i), Name: "a"}
		a = append(a, d)
	}
	for i := 0; i < 10; i++ {
		d := &BlockData{Height: int64(i + 4*i), Name: "b"}
		b = append(b, d)
	}
	for i := 0; i < 10; i++ {
		d := &BlockData{Height: int64(i + 5*i), Name: "c"}
		c = append(c, d)
	}

	s = append(s,a)
	s = append(s,b)
	s = append(s,c)
	r := mergeCandidate(s)
	for i, v := range r {
		t.Log(i, v.Name, ":", v.Height)
	}

}
func mergeCandidate(candidates [][]*BlockData) []*BlockData {
	var result = []*BlockData{}
	for i := 0; i < len(candidates); i++ {
		result = mergeDoubleCandidate(result, candidates[i])
	}
	return result

}

//两个candidate合并
func mergeDoubleCandidate(a, b []*BlockData) []*BlockData {
	var i, j, t int = 0, 0, 0
	var tmpCandidate = make([]*BlockData, len(a)+len(b))
	for i < len(a) && j < len(b) {
		//把高度最小的放入tmp数组
		if a[i].Height <= b[j].Height {
			tmpCandidate[t] = a[i]
			i++
		} else {
			tmpCandidate[t] = b[j]
			j++
		}
		t++
	}
	//剩下的顺序放入数组
	for i < len(a) {
		tmpCandidate[t] = a[i]
		t++
		i++
	}
	for j < len(b) {
		tmpCandidate[t] = b[j]
		t++
		j++
	}
	//for i, v := range tmpCandidate {
	//	fmt.Println(i, ":", v.Height)
	//}
	return tmpCandidate
}

func TestStruct(t *testing.T){
	a := BlockData{Height:10,Name:"a"}
	b := BlockData{Height:11,Name:"b"}
	c := BlockData{Height:12,Name:"c"}
	d := BlockData{Height:13,Name:"d"}
	aa := []BlockData{}
	aa = append(aa,a)
	aa = append(aa,b)
	aa = append(aa,c)
	aa = append(aa,d)

	bb := []BlockData{}
	f := BlockData{Height:15,Name:"f"}
	ff := BlockData{Height:115,Name:"ff"}
	fff := BlockData{Height:1155,Name:"fff"}
	bb = append(bb,f)
	bb = append(bb,ff)
	bb = append(bb,fff)
	aa = append(aa,bb...)
	fmt.Println(aa)
}