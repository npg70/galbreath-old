package main

import (
	"fmt"
//	"strings"
	"bytes"
	"github.com/iand/gedcom"
	"io/ioutil"
	"log"
)

func dumpName(ind *gedcom.IndividualRecord) {
	fmt.Printf("[%s] %s\n", ind.Xref,ind.Name[0].Name)
}
func dumpIndividual(ind *gedcom.IndividualRecord) {
	dumpName(ind);
	for _,e := range ind.Event {
		fmt.Printf("  %s %s %s %s %s\n", e.Tag, e.Value, e.Type, e.Date, e.Place.Name);
	}
	for _,f := range ind.Family {
		fam := f.Family
		fmt.Printf(" family %s\n", fam.Xref)
		dumpName(fam.Wife)
	    for _, e := range fam.Event {
			fmt.Printf("     %s %s %s %s %s\n", e.Tag, e.Value, e.Type, e.Date, e.Place.Name);
		}
		for _,c:= range fam.Child {
			dumpName(c)
		}
	}
}
func findXref(ged *gedcom.Gedcom, xref string) *gedcom.IndividualRecord {

	for _, rec := range ged.Individual {
		if rec.Xref==xref {
			return rec
		}
	}
	return nil
}

func main() {
	data, err := ioutil.ReadFile("Galbreath.ged")
	if err != nil {
		log.Fatalf("cant read: %s", err)
	}
	d := gedcom.NewDecoder(bytes.NewReader(data))

	g, _ := d.Decode()

	rg := findXref(g, "P57")
	dumpIndividual(rg)

	/*
	for _, rec := range g.Individual {
		if len(rec.Name) == 0 {
			continue
	    }
		if strings.Contains(rec.Name[0].Name, "Galbreath") {
			fmt.Printf("%s:%s\n", rec.Xref, rec.Name[0].Name)
		}			
	}
	*/
}

