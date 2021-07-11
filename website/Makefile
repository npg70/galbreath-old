
build:
	hugo -D
serve:
	hugo -D server --port 8080
clean:
	rm -rf public

update:
	mkdir -p content/people
	cp ../*.md content/people/
	(cd content/people; go run ../../lineage.go galbreath-james-1659.md)
	go run junk.go content/people/*.md
	hugo server --port 8080

build:
	hugo -D 
