package main

import (
	"fmt"
	"github.com/federicoReghini/Blog-aggregator/internal/config"
)

func main() {
	config, _ := internal.Read()
	fmt.Println(config)

	config.SetUser("Fred")

	read, _ := internal.Read()

	fmt.Printf("After set user %+v\n", read)

}
