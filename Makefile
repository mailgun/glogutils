test: clean
	go test
coverage: clean
	gocov test | gocov report
annotate: clean
	FILENAME=$(shell uuidgen)
	gocov test  > /tmp/--go-test-server-coverage.json
	gocov annotate /tmp/--go-test-server-coverage.json $(fn)
clean:
	find -name flymake_* -delete
sloccount:
	 find . -name "*.go" -print0 | xargs -0 wc -l
