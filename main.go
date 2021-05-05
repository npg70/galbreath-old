package main

import (
	"io"
	"os"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
    "bytes"
	"flag"
	
    "github.com/yuin/goldmark"
)

type entry struct {
	name    string
	link    string
	intro   string
	vitals  map[string]string
	spouses []*spouse
	sources string

	// these are internal fields used
	// in generation of stuff
	up		*entry  // key of primary parent for lineage lines
	first   string  // first name
 	last	string  // last name
	gen		int     // generation number	
	count   int     // lineage number
}

func (e *entry) standardVital(key string) (string, string) {

	val := e.vitals[key]
	if val == "" {
		return "", ""
	}
	parts := strings.SplitN(val, ",", 2)

	// no birth place
	if len(parts) == 1 {
		return strings.TrimSpace(val), ""
	}

	// both date and place
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func (e *entry) Birth() (string,string) {
	return e.standardVital("birth")
}
func (e *entry) Death() (string, string) {
	return e.standardVital("death")
}
func (e *entry) Burial() (string, string) {
	return e.standardVital("burial")
}

func (e *entry) Bio() string {
	out := ""
	b, _ := e.Birth()
	if b != "" {
		out += "b. " + b
	}
	b, _ = e.Death()
	if b != "" {
		out += ", d. " + b
	}

	if len(e.spouses) == 0 {
		return out
	}
	if len(e.spouses) == 1 {
		out += "; m. <span class=spouse>" + e.spouses[0].name + "</span>"
		return out
	}

	out += "; "
	for i, s := range e.spouses {
		out += fmt.Sprintf(" m (%d). <span class=spouse>%s</span>", i+1, s.name)
	}	
	return out
}

// Source return source data as HTML
func (e *entry) Intro() string {
	if e.sources == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(e.intro), &buf); err != nil {
  		panic(err)
	}
	return buf.String()
}
// Source return source data as HTML
func (e *entry) Source() string {
	if e.sources == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(e.sources), &buf); err != nil {
  		panic(err)
	}
	out := buf.String()

	// style blockquote
	out = strings.ReplaceAll(out, "<blockquote>", "<blockquote class=blockquote>")
	return out
}

func (e *entry) SetName(name string) {
		name = strings.TrimSpace(name)
	    parts := strings.Fields(name)
		e.name = name
        e.last = parts[len(parts)-1]
        e.first = strings.Join(parts[:len(parts)-1], " ")
}

func (e *entry) HasChildren() bool {
	for _, s := range e.spouses {
		if s.HasChildren() {
			return true
		}
	}
	return false
}

func (e entry) String() string {
	out := fmt.Sprintf("name: %s\n", e.name)
	out += fmt.Sprintf("link: %s\n", e.link)
	out += fmt.Sprintf("intro: %s\n", e.intro)
	for _, s := range e.spouses {
		out += "\n"
		out += s.String()
	}
	return out
}

type spouse struct {
	name     string
	intro    string
	vitals   map[string]string
	kidintro string
	children []child

	first string
	last string
}

func (s *spouse) SetName(name string) {
		name = strings.TrimSpace(name)
	    parts := strings.Fields(name)
		s.name = name
        s.last = parts[len(parts)-1]
        s.first = strings.Join(parts[:len(parts)-1], " ")
}

func (s *spouse) HasChildren() bool {
	return len(s.children) > 0
}

func (s spouse) String() string {
	out := fmt.Sprintf("  name  : %s\n", s.name)
	out += fmt.Sprintf("  intro : %s\n", s.intro)
	out += fmt.Sprintf("  kids  : %s\n", s.kidintro)
	for _, c := range s.children {
		out += "    " + c.String() + "\n"
	}
	return out
}

type child struct {
	name  string
	fname string
	intro string
}

func (c child) String() string {
	return fmt.Sprintf("%q %q %s", c.name, c.fname, c.intro)
}
func (c child) Link() string {
	return filekey(c.fname)
}

func (c child) Intro() string {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(c.intro), &buf); err != nil {
  		panic(err)
	}
	s := buf.String()

	// remove paragraph wrapper
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "<p>")
	s = strings.TrimSuffix(s, "</p>")

	// replace <em>
	s = strings.ReplaceAll(s, "<em>", "<span class=married>")
	s = strings.ReplaceAll(s, "</em>", "</span>")

	s = strings.TrimSpace(s)
	return s
}
	
const (
	StateStart = iota
	StatePrimaryIntro
	StatePrimaryDict
	StateSpouse
	StateSpouseIntro
	StateSpouseDict
	StateChildren
	StateChild
	StateSource
)

func parseChild(s string) (child, error) {
	// remove leading number
	idx := strings.IndexByte(s, ' ')
	if idx == -1 {
		return child{}, fmt.Errorf("Got bogus child: %s", s)
	}
	s = strings.TrimSpace(s[idx:])

	// child does not carry-forward
	// get name, intro
	name := s
	intro := ""
	if idx = strings.IndexByte(s, ','); idx != -1 {
	name = strings.TrimSpace(s[:idx])
	intro = strings.TrimSpace(s[idx+1:])
	}

	// child carry-forward with link
	if name[0] != '[' {
		return child{
			name:  name,
			fname: "",
			intro: intro,
		}, nil
	}

	// name is a [name](link) or a [name][id]
	idx = strings.IndexByte(name, ']')
	link := name[idx+1:]
	// remove [, ]
	name = name[1:idx]

	link = link[1:]
	link = link[:len(link)-1]
	return child{
		name:  name,
		fname: link,
		intro: intro,
	}, nil
}

// filekey returns a 
func filekey(fname string) string {
	return strings.SplitN(fname, ".", 2)[0]
}

func parse(key string, source string) (*entry, error) {
	var currentSpouse *spouse
	e := &entry{}
	e.link = key
	source = strings.TrimSpace(source)
	state := StateStart
	lines := strings.Split(source, "\n")
	idx := 0
	for idx < len(lines) {
		line := lines[idx]
		switch state {
		case StateStart:
			if !strings.HasPrefix(line, "# ") {
				panic(fmt.Sprintf("no title in %q", line))
			}
			e.SetName(line[1:])
			state = StatePrimaryIntro
		case StatePrimaryIntro:
			if strings.HasPrefix(line, "- ") {
				state = StatePrimaryDict
				continue
			}
			e.intro += line
		case StatePrimaryDict:
			if strings.HasPrefix(line, "- ") {
				line = strings.TrimSpace(line[1:])
				// should be in form of "key: value"
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					if e.vitals == nil {
						e.vitals = make(map[string]string)
					}
					e.vitals[parts[0]] = strings.TrimSpace(parts[1])
				}
			}
			if strings.HasPrefix(line, "# Source") {
				state = StateSource
				idx++
				continue
			}
			if strings.HasPrefix(line, "## ") {
				state = StateSpouse
				continue
			}
		case StateSpouse:
			currentSpouse = &spouse{}
			currentSpouse.SetName(line[3:])
			e.spouses = append(e.spouses, currentSpouse)
			state = StateSpouseIntro
		case StateSpouseIntro:
			if strings.HasPrefix(line, "- ") {
				state = StateSpouseDict
				continue
			}
			if strings.HasPrefix(line, "# Source") {
				state = StateSource
				idx++
				continue
			}
			if strings.HasPrefix(line, "## ") {
				state = StateSpouse
				continue
			}
			if strings.HasPrefix(line, "### Children") {
				state = StateChildren
			}
			currentSpouse.intro += line
		case StateSpouseDict:
			if strings.HasPrefix(line, "# Source") {
				state = StateSource
				idx++
				continue
			}
			if strings.HasPrefix(line, "## ") {
				state = StateSpouse
				continue
			}
			if strings.HasPrefix(line, "### Children") {
				state = StateChildren
			}
		case StateChildren:
			if strings.HasPrefix(line, "# Source") {
				state = StateSource
				idx++
				continue
			}
			if strings.HasPrefix(line, "## ") {
				state = StateSpouse
				continue
			}
			if strings.HasPrefix(line, "1.") {
				state = StateChild
				continue
			}
			currentSpouse.kidintro += line
		case StateChild:
			if strings.HasPrefix(line, "# Source") {
				state = StateSource
				idx++
				continue
			}
			if strings.HasPrefix(line, "## ") {
				state = StateSpouse
				continue
			}

			if line != "" {
				c, err := parseChild(line)
				if err != nil {
					return nil, fmt.Errorf("in line %q: %s", line, err)
				}
				currentSpouse.children = append(currentSpouse.children, c)
			}
		case StateSource:
			e.sources += line + "\n"
		default:
			panic("unknown state")
		}
		idx++
	}

	return e, nil
}

func parseFile(fname string) (*entry, error) {
	source, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	e, err := parse(filekey(fname), string(source))

	if err != nil {
		return nil, err
	}
	return e, nil
}

type qe struct {
	up *entry
	fname string
	gen int
}

// Breadth-First descend
func descend(root string) ([]*entry,error) {
	out := []*entry{}
	roote := qe {
		up: nil,      // parent
		fname: root,  // filename/key
		gen: 1,       // generation num: easy to compute but this is simple
	}
	q := []qe{roote}
	for len(q) > 0 {
		current := q[0]
		q = q[1:]
		log.Printf("Reading %q", current.fname)
		e, err := parseFile(current.fname)
		if err != nil {
			log.Printf("File %q not found, skipping", current.fname)
			continue
			//return nil, err
		}
		e.up = current.up
		e.gen = current.gen
		out = append(out, e)
		for _, s := range e.spouses {
			for _, c := range s.children {
				if c.fname != "" {
					log.Printf("Adding to q: %q", c.fname)
					next := qe{
						up: e,
						fname: c.fname,
						gen: current.gen + 1,
					}
						
					q = append(q, next)
				}
			}
		}
	}
	return out, nil
}

func roman(d int) string {
    switch d {
        case 1: return "i."
        case 2: return "ii."
        case 3: return "iii."
        case 4: return "iv."
        case 5: return "v."
        case 6: return "vi."
        case 7: return "vii."
        case 8: return "viii."
        case 9: return "ix."
        case 10: return "x."
        case 11: return "xi."
        case 12: return "xii."
        case 13: return "xiii."
        case 14: return "xiv."
        case 15: return "xv."
        default:
            return "!!!"
    }
}

func lineage(people []*entry, out io.StringWriter) error {

	// add in registry style counter
	for idx, e := range people {
		e.count = idx+1
	}
	// make map of ID to *entry
	keymap := make(map[string]*entry, len(people))
	for _, e := range people {
		keymap[e.link] = e
	}

	out.WriteString("---\n")
	out.WriteString("title: fo\n")
	out.WriteString("---\n\n")

	gen := 0

	for _, e := range people {
		if e.gen > gen {
			if gen > 0 {
				out.WriteString(fmt.Sprintf("</div> <!-- generation-%d -->\n", gen))
			}
			gen++
        	out.WriteString(fmt.Sprintf("\n<div id=generation-%d>\n", gen))
        	out.WriteString(fmt.Sprintf("<h1>Generation %d</h1>\n", gen))
		}
		out.WriteString(fmt.Sprintf("\n<div class=person id=%q>\n", e.link))
	    out.WriteString(fmt.Sprintf("<p><span class=primary-num>%d.</span> <span class=primary>%s</span> ",
             e.count,
             e.name,
    	))

		// lineage
		if e.up != nil {
			p := e
    		chunks := []string{}
    		for p.up != nil {
				p = p.up
        		chunks = append(chunks, fmt.Sprintf("<i><a href=%q>%s</a></i><sup>%d</sup>", "#" + p.link, p.first, p.gen))
    		}
    		out.WriteString(" (" + strings.Join(chunks, ", ") + ") ")
		}
		out.WriteString("\n")	

		out.WriteString(fmt.Sprintf(" [<a class=github href=%q>GitHub</a>] ", "https://github.com/npg70/galbreath/blob/main/" + e.link + ".md"))
		// bio into/text/spouse info
		out.WriteString(e.Bio())

		out.WriteString("\n</p>\n")
		// write out additional information

		if val := e.Intro(); val != "" {
			out.WriteString(e.Intro())
		}

		// children
		if e.HasChildren() {
			out.WriteString(fmt.Sprintf("<div class=%q>\n", "children"))
			for _, s := range e.spouses {
				birthOrder := 0
				if s.HasChildren() {
					if s.kidintro != "" {
						out.WriteString("<p>" + s.kidintro + "</p>\n")
					} else {
						out.WriteString(fmt.Sprintf("<p> Children of %s and %s (%s) %s:</p>\n",
							e.first, s.first, s.last, e.last))
					}
					out.WriteString("<table>\n")
					for _, c := range s.children {
						birthOrder++
					    out.WriteString("   <tr>\n")
						out.WriteString("      <td>")
						if c.fname == "" {
							out.WriteString("&nbsp;&nbsp;")
						} else {
							ce := keymap[c.Link()]
							if ce == nil {
								return fmt.Errorf("unable to find child id: %q", c.Link())
							}	
    						out.WriteString(fmt.Sprintf("%d", ce.count))
						}
						out.WriteString("</td>\n")
    					out.WriteString(fmt.Sprintf("      <td class=%q>%s</td>\n", "birthorder", roman(birthOrder)))

						if c.fname == "" {
							out.WriteString(fmt.Sprintf("      <td><span class=%q>%s</span>",
       							 "child", c.name))
						} else {
							out.WriteString(fmt.Sprintf("      <td><a href=%q><span class=%q>%s</span></a>",
       							 "#"+ c.Link(), "child", c.name))
						}
    					out.WriteString(", " + c.Intro())
    					out.WriteString("</td>\n")
						out.WriteString("   </tr>\n")
					}
					out.WriteString("</table>\n")
				}
			}
			out.WriteString("</div> <!-- children -->\n")
		}

		//
		out.WriteString("<div class=source>\n")
		out.WriteString(e.Source())
		out.WriteString("\n</div>\n")
		out.WriteString(fmt.Sprintf("</div> <!-- id=%s -->\n", e.link))
	}
	out.WriteString(fmt.Sprintf("</div> <!-- generation-%d -->\n", gen))
	return nil
}

func main() {
	flag.Parse()
	name := flag.Arg(0)

	people, err := descend(name)
	if err != nil {
		log.Fatalf("failed: %s", err)
	}
	err = lineage(people, os.Stdout)
	if err != nil {
		log.Fatalf("lineage failed: %s", err)
	}
}
