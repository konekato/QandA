clean:
	$(RM) bin/kone-server
dev: clean 
	go fmt
	go build -o bin/