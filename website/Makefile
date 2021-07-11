
build:
	hugo -D
serve:
	hugo -D server --port 8080
clean:
	rm -rf public
	rm -rf content/people

update:
	mkdir -p content/people
	cp ../people/*.md content/people/
	(cd content/people; go run ../../lineage.go galbreath-james-1659.md)
	go run junk.go content/people/*.md
	hugo server --port 8080
