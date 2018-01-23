package main

import "fmt"



func (this *A) say () (left , right *A) {

	var tmp1 A
	tmp1.id = 1
	left = &tmp1
	var tmp2 A
	tmp2.id = 2
	right = &tmp2

	fmt.Printf("before assign -- this: %p, left: %p, right:%p\n", this ,left, right)

	this = &tmp1
	fmt.Printf("this assign -- this: %p, left: %p, right:%p\n", this ,left, right)

	*this = tmp2
	fmt.Printf("*this assign -- this: %p, left: %p, right:%p\n", this ,left, right)

	fmt.Printf("values : this: %d, left: %d, right: %d\n", this.id, left.id, right.id)

	return
}


func (this *A) SetValue()  {
	this.id = 1
	var tmp A
	tmp.id = 2
	*this = tmp
	fmt.Printf("this.id : %d\n", this.id)
}

type A struct {
	id int
}

func (this *A) SetThis()  {
	var tmp A
	this = &tmp
	fmt.Printf("this ptr addr: %p\n", this)
}

func main()  {
	var a A
	fmt.Printf("before set this, a addr; %p\n", &a)
	a.SetThis()
	fmt.Printf("after set this, a addr; %p\n", &a)
}