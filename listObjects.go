package main

//Dok represents a documents structure
type Dok struct {
	prev *Document
	next *Document
	key  interface{}
}

//ListDocument represents a list of documents in a bucket
type ListDocument struct {
	head *Dok
	tail *Dok
}

//initialize a new list of documents
func (d *Dok) init() *Dok {
	if d == nil {
		return nil
	}

	dok := &Dok{
		prev: d.prev,
		key:  d.key,
	}
	return dok
}
