package cwmpclient

import (
	"fmt"
)

type MibValue struct {
}

type Mib struct {
	Tree map[string]string
}

func NewMib() *Mib {
	fmt.Println("creating new struct")
	m := &Mib{}
	return m
}

func (mib *Mib) AddSubTree(path string) {

}

func (mib *Mib) GetValue(path string) {

}

func (mib *Mib) SetValue(path string, value string) {

}
