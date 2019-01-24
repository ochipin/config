all:
	go test -v -cover -coverprofile cover.out
	go tool cover -func=cover.out

html:
	go test -cover -coverprofile cover.out
	go tool cover -html=cover.out

clean:
	rm cover.out
