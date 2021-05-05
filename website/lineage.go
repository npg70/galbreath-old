package main

import (
	"bytes"
	"strings"
	"fmt"
	"io/ioutil"
	"flag"
	"regexp"
	"log"
)

type ancestor struct {
	file string
	gen  int
	name string
	up   string
}

var lineage map[string]ancestor

var q []ancestor

func firstname(s string) string {
	parts := strings.Fields(s)
	return strings.Join(parts[:len(parts)-1], " ")
}

func lin(a ancestor) string {
	parts := []string{}
	for a.up != "" {
		a = lineage[a.up]
		link := fmt.Sprintf("[%s](%s)<sup>%d</sup>", firstname(a.name), a.file, a.gen)
		parts = append(parts, link)
    }		
	return strings.Join(parts, ", ")
}
	
// this convert github links to other files into hugo web links:wq
func main() {
	flag.Parse()

	lineage = make(map[string]ancestor)
	q = make([]ancestor, 0, 100)

	// finds first name '**foo bar**' on page.  It's assumed
	// this is the principal person
	namere := regexp.MustCompile(`\*\*([^*]+)\*\*`)

	// carry-forward children links
	re := regexp.MustCompile(`\d+\. \[\*([^*]+)\*\]\(([^)]+)\)`)

	childre := regexp.MustCompile(`\n(\d+)\.(.*)\n\d+`)

	args := flag.Args()
	if len(args) == 0 {
		log.Fatalf("need one filename")
	}
	q = append(q, ancestor{
		file: args[0],
		gen: 1,
		})

	for {
		if len(q) == 0 {
			break
		}
		a := q[0]
		q= q[1:]

		log.Printf("reading gen %d %s", a.gen, a.file)
		orig, err := ioutil.ReadFile(a.file)
		if err != nil {
			panic(err)
		}

		// get Primary Name
		smatch := namere.FindSubmatch(orig)
		if smatch == nil {
			continue
		}
		a.name = string(smatch[1])
		lineage[a.file] = a
		chain := lin(a)

		log.Printf("    WHOLE: %s", string(smatch[0]))
		log.Printf("    name: %s -> %s", a.name, chain)
		matched := re.FindAllSubmatch(orig, -1)
		for _, m := range matched {
			next := ancestor{
				file: string(m[2]),
				gen: a.gen+1,
				up : a.file,
			}
			log.Printf("   adding %s", next.file)
			q = append(q, next)
		}
	
		// add lineage name
		orig = bytes.Replace(orig, []byte("---"), []byte("---\nLineage: Descendants of James Galbreath and Mary Nielson"), 1)

		// let's take a look at children
		kids := childre.FindAll(orig, -1)
		for _, k := range kids {
			log.Printf("    KID: %s", k)
		}
			
		if len(chain) > 0 {
			replacement := "**" + a.name + "** (<i>" + chain + "</i>)"
			orig = bytes.Replace(orig, smatch[0], []byte(replacement), 1)
		}
	    err = ioutil.WriteFile(a.file, orig, 0644)
   		if err != nil {
           panic(err)
       	} 
	}
}
