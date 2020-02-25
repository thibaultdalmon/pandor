package main

func main() {
	e := newRouter()
	e.Logger.Fatal(e.Start(":1323"))
}
