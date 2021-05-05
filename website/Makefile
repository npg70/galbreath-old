
build:
	hugo -D
serve:
	hugo -D server
clean:
	rm -rf public

update:
	git pull
	cp ../*.md content/people/
	(cd content/people; go run ../../lineage.go galbreath-james-1659.md)
	go run junk.go content/people/*.md
	hugo server

