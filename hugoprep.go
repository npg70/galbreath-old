package main

// converts various things in github markdown to 
// work with hugo.
//
// in particular links to different files
//

import (
	"bytes"
	"io/ioutil"
	"flag"
	"regexp"
	"log"
)

// this convert github links to other files into hugo web links:wq
func main() {
	flag.Parse()

	// ( Not starting with "[{<" *.md)
	re := regexp.MustCompile(`\(([^[{<][^.]+\.md)\)`)

	for _, a := range flag.Args() {
		log.Printf("reading " + a)
		orig, err := ioutil.ReadFile(a)
		if err != nil {
			panic(err)
		}
		raw := re.ReplaceAll(orig, []byte(`({{< relref "$1" >}})`))
		if bytes.Compare(raw,orig) == 0 {
			log.Printf("no change")
			continue
		}
		err = ioutil.WriteFile(a, raw, 0644)
		if err != nil {
			panic(err)
		}
	}	
}

