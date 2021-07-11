
build:
	(cd website; hugo -D)
serve:
	(cd website; hugo -D server --port 8080)
clean:
	rm -rf website/public
	rm -rf website/content/people

update:
	cp -r people website/content
	(cd website/content/people; go run ../../../lineage.go galbreath-james-1659.md)
	go run hugoprep.go website/content/people/*.md
	(cd website; hugo server --port 8080)
