package main

func main() {
	count := 0
	for i := 0; i < 1000; i++ {
		go func() {
			count = count + 1
		}()
	}
	println(count)
}